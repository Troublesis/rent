// Command seed populates the development database with realistic fixture
// data. Usage:
//
//	go run ./cmd/seed --reset           # drop + reseed using DB_PATH
//	go run ./cmd/seed --reset --db /tmp/rent-dev.db
//
// The --reset flag is required as a safety net so that the binary cannot
// accidentally wipe a production database.
package main

import (
	"flag"
	"log"
	"time"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/seed"
	"github.com/troublesis/rent/internal/storage"
)

func main() {
	reset := flag.Bool("reset", false, "drop and re-create all tables before seeding (required)")
	dbPath := flag.String("db", "", "override DB_PATH for this run")
	flag.Parse()

	if !*reset {
		log.Fatalf("refusing to run without --reset (this command will wipe the target database)")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := storage.ApplyTimezone(cfg); err != nil {
		log.Fatalf("%v", err)
	}
	if err := storage.EnsureDirs(cfg); err != nil {
		log.Fatalf("%v", err)
	}
	db, err := storage.Open(cfg, *dbPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := storage.Reset(db); err != nil {
		log.Fatalf("reset schema: %v", err)
	}
	stats, err := seed.Run(db, time.Now())
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seeded: rooms=%d tenants=%d payments=%d settings=%d", stats.Rooms, stats.Tenants, stats.Payments, stats.Settings)
}
