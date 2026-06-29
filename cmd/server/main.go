package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	router, pushScheduler := server.NewRouter(cfg, db)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	pushScheduler.Start(ctx)

	srv := &http.Server{Addr: cfg.Addr(), Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("run server: %v", err)
		}
	}()

	log.Printf("rent app listening on http://localhost:%s (addr %s)", cfg.AppPort, cfg.Addr())

	<-ctx.Done()
	log.Printf("shutdown signal received, draining connections...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Printf("rent app stopped")
}
