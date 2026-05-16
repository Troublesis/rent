# CLAUDE.md

个人房东收租系统 (Personal Landlord Rental Management System)

## Project Summary

A personal web-based rental management tool for a single landlord. Manages properties,
tenants, leases, and payment records. Has two surfaces: a password-protected admin panel
(`/admin`) for the landlord and a read-only public listing page for prospective tenants.

**UI local rules:**
- All UI-facing text must be Simplified Chinese only.
- Do not render English labels, headings, badges, placeholders, button text, chart labels, nav copy, or decorative microcopy.
- Public pages must not show admin/login links or buttons.

Built in Go as a learning project transitioning from Python/FastAPI. Keep things simple
and idiomatic — prefer the standard library before reaching for a package.

---

## Tech Stack

| Layer         | Choice               | Note                                          |
|---------------|----------------------|-----------------------------------------------|
| Language      | Go 1.22+             | Managed via `mise` + `.tool-versions`         |
| Web framework | Gin                  | `github.com/gin-gonic/gin`                    |
| ORM           | GORM                 | `gorm.io/gorm` + `gorm.io/driver/sqlite`      |
| Database      | SQLite               | Single file `data/rent.db` — zero setup            |
| Templates     | `html/template`      | Go standard library, no extra dependency      |
| Frontend      | HTMX + Tailwind CDN  | No build step, loaded from CDN in base layout |
| Auth          | Sessions via cookies | `github.com/gin-contrib/sessions`             |
| Config        | `.env` file          | `github.com/joho/godotenv`                    |

---

## Project Structure

```
my-rent-system/
│
├── cmd/
│   └── server/
│       └── main.go            # Entry point: wires up router, DB, starts server
│
├── internal/
│   ├── auth/
│   │   └── middleware.go      # Session check — redirects /admin/* if not logged in
│   ├── handler/
│   │   ├── admin_room.go      # Admin: list, create, edit, delete rooms
│   │   ├── admin_tenant.go    # Admin: check-in, check-out, tenant detail
│   │   ├── admin_payment.go   # Admin: record payments, mark paid/unpaid
│   │   ├── admin_stats.go     # Admin: income charts and summary data
│   │   ├── admin_settings.go  # Admin: app config (landlord name, contact, etc.)
│   │   ├── auth.go            # Login / logout handlers
│   │   └── public.go          # Public: listing page, room detail, contact
│   ├── model/
│   │   └── models.go          # All GORM structs (Room, Tenant, Payment, etc.)
│   ├── repository/
│   │   ├── room_repo.go       # DB queries for rooms
│   │   ├── tenant_repo.go     # DB queries for tenants
│   │   └── payment_repo.go    # DB queries for payments
│   └── service/
│       ├── room_service.go    # Business logic for rooms
│       ├── tenant_service.go  # Business logic for tenants (check-in/out flow)
│       └── payment_service.go # Business logic for billing
│
├── templates/
│   ├── layout/
│   │   ├── admin_base.html    # Admin shell: sidebar nav, head, scripts
│   │   └── public_base.html   # Public shell: top nav, head, scripts
│   ├── admin/
│   │   ├── dashboard.html
│   │   ├── rooms.html
│   │   ├── room_form.html
│   │   ├── tenants.html
│   │   ├── payments.html
│   │   ├── stats.html
│   │   └── settings.html
│   ├── auth/
│   │   └── login.html
│   └── public/
│       ├── index.html
│       ├── rooms.html
│       └── room_detail.html
│
├── static/
│   ├── css/
│   │   └── app.css            # Minimal custom overrides only; Tailwind via CDN
│   └── js/
│       └── app.js             # Tiny helpers (confirm dialogs, etc.)
│
├── data/
│   ├── rent.db                # Local SQLite database (ignored)
│   └── uploads/               # Uploaded room images/videos (ignored except .gitkeep)
│
├── config/
│   └── config.go              # Loads .env, exposes typed Config struct
│
├── .env                       # Local secrets (not committed to git)
├── .env.example               # Template showing required keys
├── .tool-versions             # mise: pins Go version (e.g., go 1.22.4)
├── .gitignore
├── go.mod
├── go.sum
└── CLAUDE.md                  # This file
```

---

## Environment Setup (Python Developer Translation)

### Go version management — use `mise` (equivalent to `uv` for Python versions)

```bash
# Install mise (one-time, replaces pyenv/asdf/goenv)
curl https://mise.run | sh
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc  # or bash

# Inside the project root:
mise use go@1.22   # creates .tool-versions, pins the version
mise install       # downloads and installs that Go version
```

### Dependency management — `go mod` (equivalent to `uv add` / `uv sync`)

```bash
# Init (already done — go.mod exists)
go mod init github.com/yourname/my-rent-system

# Add a package (like `uv add gin`)
go get github.com/gin-gonic/gin

# Sync/tidy after cloning or editing go.mod (like `uv sync`)
go mod tidy

# No virtual env activation needed. Ever.
```

### `.env` file (required, not committed)

```env
APP_PORT=8080
APP_ENV=development
SESSION_SECRET=change-this-to-a-random-32-char-string
DB_PATH=./data/rent.db
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-password-here
UPLOAD_DIR=./data/uploads
```

---

## Running the App

```bash
# Development (with live reload via air)
go install github.com/air-verse/air@latest
air

# Or just run directly
go run ./cmd/server/main.go

# Build a production binary
go build -o rent-app ./cmd/server/main.go
./rent-app

# Cross-compile for Linux server (from Mac/Windows)
GOOS=linux GOARCH=amd64 go build -o rent-app-linux ./cmd/server/main.go
```

---

## Database

SQLite — single file at `./data/rent.db`. GORM handles migrations automatically on startup
via `db.AutoMigrate(...)` in `main.go`. No migration scripts needed for Phase 1.

To inspect the database directly:

```bash
sqlite3 data/rent.db
.tables
SELECT * FROM rooms;
```

---

## Key Conventions

### Code style
- All source code, variable names, comments, and function names in **English**.
- All UI-facing text (HTML templates) in **Simplified Chinese (zh-CN)**.
- Follow standard Go formatting — run `gofmt ./...` or use `goimports`.
- Error handling: always check errors explicitly. No panic in handlers — return HTTP errors.

### Naming
- Handler functions: `VerbNounHandler` — e.g., `ListRoomsHandler`, `CreatePaymentHandler`
- Service functions: plain verbs — e.g., `CheckInTenant`, `RecordPayment`
- Repository functions: `GetX`, `ListX`, `CreateX`, `UpdateX`, `DeleteX`
- Models: singular nouns — `Room`, `Tenant`, `Payment`

### Money / currency
- All monetary values stored as **integers in cents/fen** (e.g., ¥1500.00 → `150000`).
- Never use `float64` for money. Convert to yuan only at display time in templates.
- Template helper: `{{divideBy100 .Price}}` → renders as `1500.00`.

### File uploads
- Store uploaded files under `data/uploads/{room_id}/`.
- Save the public URL path in `room_media.url` (e.g., `/uploads/1/photo.jpg`).
- Max file size: 10MB per file. Accepted types: jpg, png, mp4.

### Auth
- Admin session stored in a signed cookie using `SESSION_SECRET`.
- Successful login always stores a persistent admin session cookie; UI label is `记住密码` and must not mention a day count.
- All `/admin/*` routes protected by `auth.RequireLogin` middleware.
- Login endpoint: `POST /admin/login` — checks against `ADMIN_USERNAME` / `ADMIN_PASSWORD` in `.env`.

---

## Routes Reference

### Public (no auth)
```
GET  /                   Public homepage — featured listings
GET  /rooms              All available rooms (paginated)
GET  /room/:id           Single room detail with contact info
```

### Auth
```
GET  /admin/login        Login form
POST /admin/login        Authenticate
POST /admin/logout       Clear session
```

### Admin (session required)
```
GET  /admin/dashboard    KPI summary, recent activity log
GET  /admin/rooms        Room list with status filters
GET  /admin/rooms/new    Create room form
POST /admin/rooms        Submit new room
GET  /admin/rooms/:id    Room detail (admin view)
GET  /admin/rooms/:id/edit  Edit room form
POST /admin/rooms/:id    Update room
POST /admin/rooms/:id/delete  Delete room

GET  /admin/tenants          Tenant list
GET  /admin/tenants/new      Check-in form
POST /admin/tenants/checkin  Submit check-in
GET  /admin/tenants/:id      Tenant detail
POST /admin/tenants/:id/checkout  Process checkout

GET  /admin/payments         Payment list with filters
POST /admin/payments         Record a new payment
POST /admin/payments/:id/toggle  Mark paid / unpaid

GET  /admin/stats            Income analytics charts
GET  /admin/settings         App settings form
POST /admin/settings         Save settings
```

### API (JSON, used by HTMX)
```
POST /api/upload             Upload room images/videos (multipart)
GET  /api/dashboard/stats    JSON stats for chart rendering
```

---

## Database Schema (GORM Models)

### Room
```go
type Room struct {
    ID          uint      `gorm:"primarykey"`
    RoomNo      string    `gorm:"uniqueIndex;not null"`   // e.g. "A101"
    Title       string    `gorm:"not null"`
    Description string
    Price       int       // cents — e.g. 150000 = ¥1500.00
    Deposit     int       // cents
    Status      string    `gorm:"default:vacant"`         // vacant | occupied | maintenance
    Area        int       // square meters
    Floor       int
    Tags        string    // comma-separated: "电梯,阳台,空调"
    Media       []RoomMedia
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Tenant
```go
type Tenant struct {
    ID               uint      `gorm:"primarykey"`
    Name             string    `gorm:"not null"`
    Phone            string    `gorm:"not null"`
    EmergencyContact string
    RoomID           uint      `gorm:"not null"`
    Room             Room
    CheckinDate      time.Time
    CheckoutDate     *time.Time  // nil while still active
    RentPrice        int         // agreed monthly rent in cents
    Deposit          int         // collected deposit in cents
    Status           string      `gorm:"default:active"` // active | checkout
    Payments         []Payment
    CreatedAt        time.Time
}
```

### Payment
```go
type Payment struct {
    ID       uint      `gorm:"primarykey"`
    TenantID uint      `gorm:"not null"`
    Amount   int       // cents
    Type     string    // rent | water | electricity | other
    Paid     bool      `gorm:"default:false"`
    PayDate  time.Time
    Note     string
    CreatedAt time.Time
}
```

---

## Phase 1 MVP Scope (Build in this order)

1. **Project bootstrap** — `main.go`, DB connection, GORM AutoMigrate, Gin router
2. **Auth** — login/logout, session middleware, login page
3. **Rooms CRUD** — list, create, edit, delete, image upload
4. **Tenants** — check-in form, tenant list, checkout flow
5. **Payments** — record payment, toggle paid/unpaid, list with filters
6. **Dashboard** — KPI counts (rooms, tenants, arrears, monthly income)
7. **Public listing page** — room cards, room detail, contact display
8. **Stats page** — simple monthly income bar chart (use Chart.js CDN)

## Phase 2 (after MVP works)
- WeChat Pay / Alipay integration
- Auto-generate utility bills from meter readings
- SMS reminders (Aliyun SMS)
- PDF lease agreement export
- PWA manifest for mobile home screen install
- Swagger/OpenAPI docs (add `swag` annotations at this point)

---

## Common Commands Cheatsheet

```bash
mise install              # Install pinned Go version
go mod tidy               # Sync dependencies (like uv sync)
go run ./cmd/server/main.go  # Run locally
air                       # Run with live reload (install: go install github.com/air-verse/air@latest)
gofmt -w ./...            # Format all Go files
go vet ./...              # Lint check
go test ./...             # Run tests
go build -o rent-app ./cmd/server/main.go  # Build binary
```

---

## What NOT to do (for Phase 1)

- Do not add Swagger/swag annotations yet — adds noise before you understand the codebase
- Do not use PostgreSQL yet — SQLite is fine for one landlord, swap when you need it
- Do not add Redis or any cache layer — premature for this scale
- Do not use goroutines in handlers yet — learn the sync patterns first
- Do not split into microservices — this is a monolith by design
