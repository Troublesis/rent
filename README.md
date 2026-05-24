# 个人房东收租系统 (Rent)

A personal web-based rental management tool for a single landlord — built in Go with Gin, GORM and SQLite. Includes a password-protected admin panel (`/admin`) and a read-only public listing page.

For project design, conventions, and scope, see [CLAUDE.md](CLAUDE.md).

---

## Requirements

| Tool   | Version  | Notes                                                |
|--------|----------|------------------------------------------------------|
| Go     | 1.25.0   | Pinned in [.tool-versions](.tool-versions)           |
| mise   | latest   | Recommended for Go version management                |
| SQLite | bundled  | Pure-Go driver, no system SQLite required            |
| air    | optional | Live-reload during development                       |

---

## Setup

### 1. Install Go via mise

```bash
curl https://mise.run | sh
eval "$(mise activate zsh)"   # or bash

mise install                  # installs Go 1.25.0 from .tool-versions
```

### 2. Clone and sync dependencies

```bash
git clone git@github.com:troublesis/rent.git
cd rent
go mod tidy
```

### 3. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set at minimum:

| Key                | Purpose                                  |
|--------------------|------------------------------------------|
| `APP_PORT`         | HTTP listen port (default `8080`)        |
| `SESSION_SECRET`   | Random 32+ char string, signs cookies    |
| `DB_PATH`          | SQLite file path, default `./data/rent.db` |
| `ADMIN_USERNAME`   | Admin login username                     |
| `ADMIN_PASSWORD`   | Admin login password                     |
| `UPLOAD_DIR`       | Path for uploaded room media             |
| `LANDLORD_NAME`    | Display name on public page              |
| `LANDLORD_PHONE`   | Contact phone on public page             |

The `data/` and `data/uploads/` directories are created automatically on first run.

---

## Run

### Development

```bash
go run ./cmd/server
```

Then open:

- Public listing: <http://localhost:8080/>
- Admin panel:   <http://localhost:8080/admin/login>

### Live reload (optional)

```bash
go install github.com/air-verse/air@latest
air
```

Air is driven by [`.air.toml`](.air.toml) at the repo root, which points the build
at `./cmd/server` (the actual entry point). Without that config, `air` defaults to
`go build .` and fails with `no Go files in <repo>`.

### Production build

```bash
# Local binary
go build -o rent-app ./cmd/server
./rent-app

# Cross-compile for a Linux server
GOOS=linux GOARCH=amd64 go build -o rent-app-linux ./cmd/server
```

The binary is self-contained — copy `rent-app`, `.env`, `templates/`, `static/`, and `data/` to the target host.

---

## Seeding the Database

Two seeders live under `cmd/`. Both require `--reset` as a safety net — it drops and
re-creates all tables against the target DB before inserting fixtures.

```bash
# Small realistic dev dataset (default DB_PATH from .env)
go run ./cmd/seed --reset

# Same, but against a throwaway DB
go run ./cmd/seed --reset --db /tmp/rent-dev.db

# Bulk dataset for performance / UI stress-testing (~1200 rooms by default)
go run ./cmd/seed-bulk --reset

# Custom room count + isolated DB + deterministic RNG seed
go run ./cmd/seed-bulk --reset --rooms 2000
go run ./cmd/seed-bulk --reset --db /tmp/perf.db --rooms 5000 --seed 123
```

`seed-bulk` flags:

| Flag      | Default | Purpose                                         |
|-----------|---------|-------------------------------------------------|
| `--reset` | (req'd) | Drop + re-create all tables before seeding     |
| `--rooms` | 1200    | Approximate number of rooms to generate         |
| `--db`    | (env)   | Override `DB_PATH` for this run                 |
| `--seed`  | 42      | RNG seed for deterministic output               |

> Warning: never point `--db` at a database you care about — `--reset` wipes it.

---

## CLI Cheatsheet

```bash
# Dependencies
go mod tidy                   # Sync deps (equivalent to `uv sync`)
go get <pkg>                  # Add a dependency

# Run
go run ./cmd/server           # Start dev server
air                           # Start with live reload

# Seed data
go run ./cmd/seed --reset                    # Small realistic dev dataset
go run ./cmd/seed-bulk --reset --rooms 2000  # Bulk dataset for perf/UI testing

# Build
go build -o rent-app ./cmd/server
GOOS=linux GOARCH=amd64 go build -o rent-app-linux ./cmd/server

# Quality
gofmt -w ./...                # Format all Go files
go vet ./...                  # Static analysis

# Tests
go test ./...                                   # Run all tests
go test ./internal/handler/...                  # Test one package
go test -run TestRecordPayment ./internal/...   # Run by name
go test -v ./...                                # Verbose
go test -cover ./...                            # With coverage summary
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
go test -race ./...                             # Race detector
```

---

## Testing

Tests live alongside the code they cover:

```
config/config_test.go
internal/handler/*_test.go        # HTTP handler tests
internal/repository/*_test.go     # GORM/SQLite repo tests
internal/seed/seed_test.go
internal/server/*_test.go         # Router + template rendering
```

Repository and handler tests use an in-memory / temp-file SQLite DB — no external services required. Just:

```bash
go test ./...
```

---

## Database

SQLite single-file at `./data/rent.db`. GORM `AutoMigrate` runs on every startup, so schema changes in [internal/model](internal/model) take effect immediately.

Inspect directly:

```bash
sqlite3 data/rent.db
.tables
SELECT * FROM rooms;
```

To reset the DB during development:

```bash
rm data/rent.db
go run ./cmd/server   # recreated empty on next start
```

---

## Project Layout

```
cmd/server/         # main.go entry point
config/             # .env loader, typed Config
internal/
  auth/             # session middleware
  handler/          # HTTP handlers (admin_*, public, auth)
  model/            # GORM models
  repository/       # DB queries
  service/          # Business logic
  seed/             # Seed data helpers
  server/           # Router wiring, template registry
templates/          # html/template files (admin, public, auth, layout)
static/             # CSS / JS served at /static
data/               # SQLite DB + uploaded media (gitignored)
```

See [CLAUDE.md](CLAUDE.md) for full route reference, schema, and conventions.

---

## Makefile

A `Makefile` is included for common tasks. Run `make` (or `make help`) to list all targets:

```
  build          Format, vet, then build the production binary
  build-linux    Cross-compile for Linux amd64
  db             Open SQLite shell on the app database
  db-reset       Delete the local dev database (next run recreates it)
  deploy         Build binary and restart the systemd service
  dev            Start dev server with live reload
  fmt            Format all Go source files
  logs           Tail rent app logs (Ctrl-C to exit)
  logs-caddy     Tail Caddy logs (Ctrl-C to exit)
  reload-caddy   Reload Caddy config without downtime
  restart        Restart the systemd service
  run            Run dev server directly (no live reload)
  seed           Seed realistic dev data (drops and recreates DB)
  seed-bulk      Seed bulk data (~1200 rooms) for perf/UI testing
  start          Start the systemd service
  status         Show status of rent + caddy services
  stop           Stop the systemd service
  test           Run all tests
  test-cover     Run tests and open HTML coverage report
  test-v         Run all tests (verbose)
  vet            Run go vet static analysis
```

---

## Production & Operations

The production setup uses Caddy as a reverse proxy with automatic HTTPS (Let's Encrypt),
and the Go app runs as a systemd service.

### Stack

| Component     | Detail                                   |
|---------------|------------------------------------------|
| Domain        | `15158920228.xyz` — A record → this host |
| TLS           | Let's Encrypt via Caddy (auto-renews)    |
| Reverse proxy | Caddy `/etc/caddy/Caddyfile`             |
| App process   | `systemd` unit `rent.service`            |
| App binding   | `127.0.0.1:8080` (localhost only)        |

### Daily operations

```bash
# Service status (rent + caddy)
make status

# Live log stream
make logs          # app logs
make logs-caddy    # Caddy / TLS logs

# Restart app (e.g. after config change)
make restart

# Deploy a new build
make deploy        # = go build + systemctl restart rent

# Reload Caddy (after editing /etc/caddy/Caddyfile)
make reload-caddy
```

### Editing Caddy config

```bash
sudoedit /etc/caddy/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile   # validate first
make reload-caddy
```

### Rotating secrets

```bash
openssl rand -hex 32                 # generate a new SESSION_SECRET
sudoedit /home/opc/code/rent/.env    # update SESSION_SECRET (and ADMIN_PASSWORD)
make restart                         # pick up the new values
```

> Note: rotating `SESSION_SECRET` invalidates all existing login sessions.

### Cert troubleshooting

Caddy auto-renews — nothing to do under normal operation. To force-check:

```bash
make logs-caddy          # look for renewal or error messages
sudo caddy validate --config /etc/caddy/Caddyfile
```

---

## Troubleshooting

| Symptom                                  | Fix                                                                 |
|------------------------------------------|---------------------------------------------------------------------|
| `SESSION_SECRET` warning on startup      | Set a 32+ char random value in `.env`                               |
| Admin login rejects valid credentials    | Check `ADMIN_USERNAME` / `ADMIN_PASSWORD` in `.env`, restart server |
| `no such table` errors                   | Delete `data/rent.db` and restart to re-run AutoMigrate             |
| Uploads return 404                       | Ensure `UPLOAD_DIR` exists and matches the value in `.env`          |
| Port already in use                      | Change `APP_PORT` in `.env`                                         |
