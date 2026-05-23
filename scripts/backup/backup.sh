#!/usr/bin/env bash
# Rent app SQLite + .env backup → WebDAV with rotation and Bark notifications.
# See README.md in this directory for setup and usage.
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${BACKUP_ENV_FILE:-${SCRIPT_DIR}/.env}"

# ---------- logging ----------
log() {
  printf '%s [backup] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >&2
  if [[ -n "${LOG_FILE:-}" ]]; then
    printf '%s [backup] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >>"$LOG_FILE" || true
  fi
}

die() {
  log "FATAL: $*"
  notify_failure "$*" || true
  exit 1
}

# ---------- bark ----------
bark_post() {
  local title="$1" body="$2" level="${3:-active}"
  [[ -z "${BARK_URL:-}" || -z "${BARK_KEY:-}" ]] && return 0
  local url="${BARK_URL%/}/${BARK_KEY}"
  local payload
  payload=$(cat <<EOF
{"title":$(json_str "$title"),"body":$(json_str "$body"),"group":$(json_str "${BARK_GROUP:-rent-backup}"),"level":$(json_str "$level"),"icon":$(json_str "${BARK_ICON:-}")}
EOF
)
  curl -sS -o /dev/null \
    --connect-timeout 10 --max-time 15 --retry 2 --retry-delay 2 \
    -H 'Content-Type: application/json; charset=utf-8' \
    -X POST -d "$payload" "$url" || log "WARN: bark notify failed (non-fatal)"
}

json_str() {
  local s="${1:-}"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  printf '"%s"' "$s"
}

notify_success() {
  [[ "${BARK_NOTIFY_ON_SUCCESS:-1}" == "1" ]] || return 0
  bark_post "✅ Rent backup OK" "$1" "active"
}

notify_failure() {
  bark_post "❌ Rent backup FAILED" "$1" "timeSensitive"
}

# ---------- load env ----------
[[ -f "$ENV_FILE" ]] || { echo "Missing env file: $ENV_FILE" >&2; exit 1; }
# shellcheck source=/dev/null
set -a; . "$ENV_FILE"; set +a

# ---------- validate ----------
: "${DB_PATH:?DB_PATH not set}"
: "${WEBDAV_URL:?WEBDAV_URL not set}"
: "${WEBDAV_USER:?WEBDAV_USER not set}"
: "${WEBDAV_PASS:?WEBDAV_PASS not set}"
BACKUP_KEEP_N="${BACKUP_KEEP_N:-30}"
BACKUP_PREFIX="${BACKUP_PREFIX:-rent}"
UPLOAD_TIMEOUT="${UPLOAD_TIMEOUT:-300}"
LOCK_FILE="${LOCK_FILE:-/tmp/rent-backup.lock}"
STATE_FILE="${STATE_FILE:-${SCRIPT_DIR}/.last-backup-hash}"
APP_ENV_FILE="${APP_ENV_FILE-}"
[[ "$WEBDAV_URL" == */ ]] || WEBDAV_URL="${WEBDAV_URL}/"

# ---------- args ----------
DRY_RUN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    -h|--help) sed -n '2,4p' "$0"; exit 0 ;;
    *) die "Unknown argument: $arg" ;;
  esac
done

# ---------- deps ----------
for bin in sqlite3 curl gzip tar; do
  command -v "$bin" >/dev/null 2>&1 || die "Missing required binary: $bin"
done

if command -v sha256sum >/dev/null 2>&1; then
  HASH_CMD=(sha256sum)
elif command -v shasum >/dev/null 2>&1; then
  HASH_CMD=(shasum -a 256)
else
  die "Missing required binary: sha256sum or shasum"
fi

hash_file() {
  "${HASH_CMD[@]}" "$1" | awk '{print $1}'
}

# ---------- single-instance lock ----------
if command -v flock >/dev/null 2>&1; then
  exec 9>"$LOCK_FILE"
  flock -n 9 || die "Another backup run is in progress (lock: $LOCK_FILE)"
fi

# ---------- workspace ----------
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARCHIVE_NAME="${BACKUP_PREFIX}-${STAMP}.tar.gz"
ARCHIVE_PATH="${TMP}/${ARCHIVE_NAME}"

# ---------- snapshot db ----------
[[ -f "$DB_PATH" ]] || die "DB_PATH not found: $DB_PATH"
log "Snapshotting SQLite DB from $DB_PATH"
sqlite3 "$DB_PATH" ".backup '${TMP}/rent.db'" || die "sqlite3 .backup failed"
sqlite3 "${TMP}/rent.db" 'PRAGMA integrity_check;' | head -1 | grep -q '^ok$' \
  || die "Snapshot failed integrity_check"

# ---------- capture env ----------
ENV_INCLUDED="no"
if [[ -n "$APP_ENV_FILE" ]]; then
  [[ -f "$APP_ENV_FILE" ]] || die "APP_ENV_FILE set but not found: $APP_ENV_FILE"
  cp "$APP_ENV_FILE" "${TMP}/app.env"
  chmod 600 "${TMP}/app.env"
  ENV_INCLUDED="yes"
  log "Included app env file from $APP_ENV_FILE"
else
  log "WARN: APP_ENV_FILE empty — archive will NOT contain the app .env"
fi

# ---------- change detection ----------
DB_HASH="$(hash_file "${TMP}/rent.db")"
ENV_HASH=""
if [[ "$ENV_INCLUDED" == "yes" ]]; then
  ENV_HASH="$(hash_file "${TMP}/app.env")"
fi
CURRENT_HASH="${DB_HASH}:${ENV_HASH}"

if [[ -f "$STATE_FILE" ]]; then
  LAST_HASH="$(cat "$STATE_FILE" 2>/dev/null || true)"
else
  LAST_HASH=""
fi

if [[ -n "$LAST_HASH" && "$LAST_HASH" == "$CURRENT_HASH" ]]; then
  log "No changes since last backup (hash=${DB_HASH%%:*}…) — skipping upload and rotation."
  exit 0
fi

# ---------- bundle ----------
log "Creating archive $ARCHIVE_NAME"
if [[ "$ENV_INCLUDED" == "yes" ]]; then
  tar -C "$TMP" -czf "$ARCHIVE_PATH" rent.db app.env
else
  tar -C "$TMP" -czf "$ARCHIVE_PATH" rent.db
fi
ARCHIVE_SIZE="$(du -h "$ARCHIVE_PATH" | awk '{print $1}')"
log "Archive size: $ARCHIVE_SIZE"

# ---------- upload ----------
if [[ "$DRY_RUN" == "1" ]]; then
  log "[dry-run] Would upload $ARCHIVE_NAME to ${WEBDAV_URL}${ARCHIVE_NAME}"
else
  log "Uploading to ${WEBDAV_URL}${ARCHIVE_NAME}"
  curl -fsS --user "${WEBDAV_USER}:${WEBDAV_PASS}" \
    --connect-timeout 10 --max-time "$UPLOAD_TIMEOUT" \
    --retry 3 --retry-delay 5 \
    -T "$ARCHIVE_PATH" "${WEBDAV_URL}${ARCHIVE_NAME}" \
    || die "WebDAV upload failed"
  printf '%s\n' "$CURRENT_HASH" >"$STATE_FILE" || log "WARN: failed to update state file $STATE_FILE"
fi

# ---------- rotation ----------
log "Listing remote backups for rotation (keep last ${BACKUP_KEEP_N})"
PROPFIND_BODY='<?xml version="1.0"?><d:propfind xmlns:d="DAV:"><d:prop><d:displayname/></d:prop></d:propfind>'
LISTING="$(curl -fsS --user "${WEBDAV_USER}:${WEBDAV_PASS}" \
  --connect-timeout 10 --max-time 60 \
  -X PROPFIND -H 'Depth: 1' -H 'Content-Type: application/xml' \
  --data "$PROPFIND_BODY" "$WEBDAV_URL" 2>/dev/null)" || {
  log "WARN: PROPFIND failed — skipping rotation"
  LISTING=""
}

REMOTE_FILES=()
while IFS= read -r line; do
  [[ -n "$line" ]] && REMOTE_FILES+=("$line")
done < <(
  printf '%s' "$LISTING" \
    | grep -oE '<(d|D):href>[^<]+</(d|D):href>' \
    | sed -E 's#</?[dD]:href>##g' \
    | awk -F/ '{print $NF}' \
    | grep -E "^${BACKUP_PREFIX}-[0-9]{8}T[0-9]{6}Z\.tar\.gz$" \
    | sort -r || true
)

DELETED=0
TOTAL=${#REMOTE_FILES[@]}
if (( TOTAL > BACKUP_KEEP_N )); then
  for (( i = BACKUP_KEEP_N; i < TOTAL; i++ )); do
    old="${REMOTE_FILES[$i]}"
    if [[ "$DRY_RUN" == "1" ]]; then
      log "[dry-run] Would DELETE ${WEBDAV_URL}${old}"
    else
      log "Deleting old backup: $old"
      curl -fsS --user "${WEBDAV_USER}:${WEBDAV_PASS}" \
        --connect-timeout 10 --max-time 30 --retry 2 --retry-delay 3 \
        -X DELETE "${WEBDAV_URL}${old}" \
        || log "WARN: failed to delete $old (non-fatal)"
    fi
    DELETED=$((DELETED + 1))
  done
fi

REMAINING=$(( ${#REMOTE_FILES[@]} - DELETED + ( DRY_RUN == 1 ? 0 : 1 ) ))
(( REMAINING < 0 )) && REMAINING=0

SUMMARY="archive=${ARCHIVE_NAME} size=${ARCHIVE_SIZE} env=${ENV_INCLUDED} kept=${REMAINING} pruned=${DELETED}"
log "Done. $SUMMARY"

if [[ "$DRY_RUN" == "1" ]]; then
  log "Dry run complete — no notification sent."
else
  notify_success "$SUMMARY"
fi
