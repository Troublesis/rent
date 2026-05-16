package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/server"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("create upload directory: %v", err)
	}
	if dbDir := filepath.Dir(cfg.DBPath); dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("create database directory: %v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.Room{}, &model.RoomMedia{}, &model.Tenant{}, &model.Payment{}, &model.AppSetting{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	router := server.NewRouter(cfg, db)
	log.Printf("rent app listening on http://localhost%s", cfg.Addr())
	if err := router.Run(cfg.Addr()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
