BINARY  := rent-app
CMD     := ./cmd/server
SERVICE := rent

.PHONY: help dev build run stop restart logs status fmt vet test test-cover seed seed-bulk

# ── default ────────────────────────────────────────────────────────────────────

help:
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' | sort

# ── development ────────────────────────────────────────────────────────────────

dev: ## Start dev server with live reload (requires: go install github.com/air-verse/air@latest)
	air

run: ## Run dev server directly (no live reload)
	go run $(CMD)

fmt: ## Format all Go source files
	gofmt -w ./...

vet: ## Run go vet static analysis
	go vet ./...

test: ## Run all tests
	go test ./...

test-v: ## Run all tests (verbose)
	go test -v ./...

test-cover: ## Run tests and open HTML coverage report
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out

seed: ## Seed realistic dev data (drops and recreates DB)
	go run ./cmd/seed --reset

seed-bulk: ## Seed bulk data (~1200 rooms) for perf/UI testing
	go run ./cmd/seed-bulk --reset

# ── production build ───────────────────────────────────────────────────────────

build: fmt vet ## Format, vet, then build the production binary
	go build -o $(BINARY) $(CMD)

build-linux: ## Cross-compile for Linux amd64
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux $(CMD)

# ── production service ─────────────────────────────────────────────────────────

deploy: build ## Build binary and restart the systemd service
	sudo systemctl restart $(SERVICE)

start: ## Start the systemd service
	sudo systemctl start $(SERVICE)

stop: ## Stop the systemd service
	sudo systemctl stop $(SERVICE)

restart: ## Restart the systemd service
	sudo systemctl restart $(SERVICE)

status: ## Show status of rent + caddy services
	@echo "=== rent ===" && sudo systemctl status $(SERVICE) --no-pager -l | head -20
	@echo
	@echo "=== caddy ===" && sudo systemctl status caddy --no-pager -l | head -20

logs: ## Tail rent app logs (Ctrl-C to exit)
	sudo journalctl -u $(SERVICE) -f

logs-caddy: ## Tail Caddy logs (Ctrl-C to exit)
	sudo journalctl -u caddy -f

reload-caddy: ## Reload Caddy config without downtime
	sudo systemctl reload caddy

# ── database ───────────────────────────────────────────────────────────────────

db: ## Open SQLite shell on the app database
	sqlite3 data/rent.db

db-reset: ## Delete the local dev database (next run recreates it)
	@read -p "Delete data/rent.db? [y/N] " ok && [ "$$ok" = "y" ] && rm -f data/rent.db && echo "deleted" || echo "aborted"
