// Command seed-bulk populates the database with a large volume of realistic
// fixture data for performance and UI stress-testing. Usage:
//
//	go run ./cmd/seed-bulk --reset                     # default: ~1200 rooms
//	go run ./cmd/seed-bulk --reset --rooms 2000
//	go run ./cmd/seed-bulk --reset --db /tmp/perf.db --seed 123
//
// The --reset flag is required as a safety net.
package main

import (
	"flag"
	"log"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/seed"
	"github.com/troublesis/rent/internal/storage"
)

func main() {
	reset := flag.Bool("reset", false, "drop and re-create all tables before seeding (required)")
	dbPath := flag.String("db", "", "override DB_PATH for this run")
	roomCount := flag.Int("rooms", 1200, "approximate number of rooms to generate")
	seedVal := flag.Int64("seed", 42, "RNG seed for deterministic output")
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

	stats, err := seed.RunBulk(db, *roomCount, *seedVal)
	if err != nil {
		log.Fatalf("seed-bulk: %v", err)
	}
	log.Printf("seeded: rooms=%d tenants=%d payments=%d settings=%d",
		stats.Rooms, stats.Tenants, stats.Payments, stats.Settings)
}
