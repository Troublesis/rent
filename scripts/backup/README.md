# Rent Backup Script

A tiny, cron-friendly Bash script that snapshots the rent app's SQLite database **and** its runtime `.env`, uploads the bundle to a WebDAV target, rotates old archives, and pushes a [Bark](https://github.com/Finb/Bark) notification on success or failure.

```
   ┌────────────────┐    sqlite3 .backup     ┌────────────┐
   │ data/rent.db   │ ─────────────────────▶ │  rent.db   │
   └────────────────┘                        │            │
                                             │  app.env   │
   ┌────────────────┐    cp                  │            │  tar -czf
   │ /app/.env      │ ─────────────────────▶ │ (tmp dir)  │ ──────────▶ rent-<UTC>.tar.gz
   └────────────────┘                        └────────────┘                     │
                                                                                │ curl -T
                                                                                ▼
                                                                         ┌─────────────┐
                                                                         │   WebDAV    │
                                                                         └─────────────┘
                                                                                │
                                                                  PROPFIND + DELETE (rotation)
                                                                                │
                                                                                ▼
                                                                          📱 Bark push
```

---

## Requirements

| Tool      | Purpose                                  | Notes                     |
|-----------|------------------------------------------|---------------------------|
| `bash`    | Script runtime                           | 4+ recommended            |
| `sqlite3` | Consistent DB snapshot (`.backup`)       | Required                  |
| `curl`    | WebDAV upload / PROPFIND / DELETE + Bark | 7.x+                      |
| `tar`     | Bundle the snapshot + env                | GNU or BSD tar both OK    |
| `gzip`    | Compression (invoked by `tar -z`)        | Required                  |
| `flock`   | Single-instance lock                     | Optional but recommended  |

Quick check:

```bash
command -v bash sqlite3 curl tar gzip
```

---

## Setup

From the repo root:

```bash
cd scripts/backup
cp .env.example .env
chmod 600 .env
chmod +x backup.sh
```

Edit `.env` and fill in:

- `DB_PATH` — absolute path to `data/rent.db`.
- `APP_ENV_FILE` — absolute path to the rent app's `.env` (leave empty to skip, **not recommended**).
- `WEBDAV_URL` — directory URL **ending in `/`**. The directory must exist on the server (the script does **not** create it).
- `WEBDAV_USER` / `WEBDAV_PASS` — credentials for the WebDAV account.
- `BARK_URL` + `BARK_KEY` — server base URL (`https://api.day.app` for the hosted service) and your device key.
- `BACKUP_KEEP_N` — how many archives to retain on the remote.

### Pre-create the WebDAV directory

Most servers won't auto-create nested folders. Do this once:

```bash
# Nextcloud / dufs / generic WebDAV
curl --user "$WEBDAV_USER:$WEBDAV_PASS" -X MKCOL "$WEBDAV_URL"
```

### Bark setup

1. Install the [Bark iOS app](https://apps.apple.com/app/bark-customed-notifications/id1403753865) (or use a [self-hosted server](https://github.com/Finb/bark-server)).
2. Copy your device key from the app and paste it into `BARK_KEY`.
3. `BARK_URL` is `https://api.day.app` for the hosted service, or your own server's base URL.

---

## Usage

### Dry run (no network writes)

```bash
./backup.sh --dry-run
```

Prints the archive name and rotation plan; does **not** upload, delete, or notify.

### Real run

```bash
./backup.sh
```

Logs go to stderr. To also append to a file, set `LOG_FILE` in `.env`.

---

## Cron

```cron
# Daily at 03:30 server-local time
30 3 * * * /absolute/path/to/rent/scripts/backup/backup.sh >> /var/log/rent-backup.log 2>&1
```

Notes:

- Cron's `PATH` is minimal. If `sqlite3`/`curl` live in non-standard places, set `PATH=` at the top of the crontab.
- `MAILTO=` at the top of the crontab will email any stderr output — useful as a secondary failure channel beyond Bark.
- The script takes a `flock` lock at `LOCK_FILE`, so overlapping cron fires are safe.

To verify the cron entry is healthy:

```bash
# Schedule a one-off run 2 minutes from now and watch the log
*/2 * * * * /absolute/path/to/rent/scripts/backup/backup.sh >> /tmp/rent-backup-smoke.log 2>&1
tail -f /tmp/rent-backup-smoke.log
```

Then remove the line once you've confirmed it ran.

---

## Restoring from a backup

1. Download the most recent archive from your WebDAV target:

   ```bash
   curl -u "$WEBDAV_USER:$WEBDAV_PASS" -O "${WEBDAV_URL}rent-<UTC-stamp>.tar.gz"
   ```

2. Extract it to a temp directory:

   ```bash
   mkdir restore && tar -xzf rent-*.tar.gz -C restore
   ls restore   # → rent.db   app.env (if env was included)
   ```

3. Stop the running rent server.

4. Replace the live files:

   ```bash
   cp restore/rent.db /path/to/rent/data/rent.db
   cp restore/app.env /path/to/rent/.env   # if archive included it
   chmod 600 /path/to/rent/.env
   ```

5. Start the server. Check `/admin/dashboard` to confirm data is intact.

Alternative for online restore (server still running): `sqlite3 /path/to/rent/data/rent.db ".restore restore/rent.db"`.

---

## Troubleshooting

| Symptom                                    | Likely cause / fix                                                                              |
|--------------------------------------------|-------------------------------------------------------------------------------------------------|
| `401 Unauthorized` on upload               | Wrong `WEBDAV_USER` / `WEBDAV_PASS`, or app password required (Nextcloud).                      |
| `404 Not Found` on upload                  | `WEBDAV_URL` directory doesn't exist — run the `MKCOL` step above.                              |
| `405 Method Not Allowed` on PROPFIND       | Server doesn't allow PROPFIND for this user — rotation will be skipped, uploads still work.     |
| Bark push never arrives                    | Check `BARK_URL` has no trailing junk, `BARK_KEY` is the device key, app foreground permission. |
| `sqlite3: command not found`               | Install it: `apt install sqlite3` / `brew install sqlite`. Cron may need `PATH=` set.           |
| Rotation never deletes anything            | Filenames must match `<BACKUP_PREFIX>-YYYYMMDDTHHMMSSZ.tar.gz` — don't rename remote files.     |
| Script silently does nothing under cron    | Lock file held: `rm /tmp/rent-backup.lock`. Or `PATH` issue: log the env at the top.            |
| `Snapshot failed integrity_check`          | Source DB is corrupt — investigate with `sqlite3 data/rent.db 'PRAGMA integrity_check;'`.       |

---

## Security notes

- **The archive contains the app's `.env`** (admin password, `SESSION_SECRET`, WebDAV creds themselves). Treat the WebDAV target as sensitive. Use a dedicated user scoped to this folder if your server supports it.
- Keep `scripts/backup/.env` at `chmod 600`. Never commit it (the local `.gitignore` covers this, and the repo-root `.gitignore` already ignores `.env`).
- The script does not yet encrypt archives at rest. If your WebDAV target is shared or untrusted, layer `age` or `gpg` between `tar` and `curl` — this is a planned follow-up; see *Limitations* below.

---

## Limitations / non-goals

- Does **not** back up `data/uploads/` (room images/videos). Use a separate `rsync`/`rclone` job for that.
- Does **not** encrypt archives at rest. Mitigate via WebDAV access control until a `BACKUP_ENCRYPT_RECIPIENT` hook is added.
- Does **not** create the remote directory (`MKCOL`). One-time manual setup.
- Does **not** run as a daemon. Drive it from cron / systemd timer.
