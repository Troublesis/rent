package main

import (
	"log"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/server"
	"github.com/troublesis/rent/internal/storage"
)

func main() {
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
	db, err := storage.Open(cfg, "")
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := storage.Migrate(db); err != nil {
		log.Fatalf("%v", err)
	}

	router := server.NewRouter(cfg, db)
	log.Printf("rent app listening on http://localhost%s", cfg.Addr())
	if err := router.Run(cfg.Addr()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
