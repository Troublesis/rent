// Package storage centralizes database + filesystem bootstrap so that the
// server binary and the seed binary share identical initialization logic.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AllModels enumerates every GORM model managed by the application. Both
// AutoMigrate and the seed reset path consume this list.
func AllModels() []interface{} {
	return []interface{}{
		&model.Room{},
		&model.RoomMedia{},
		&model.Tenant{},
		&model.Payment{},
		&model.AppSetting{},
	}
}

// ApplyTimezone loads the timezone from configuration and sets it as the
// process default. Called from every entry point so that time.Now() is
// consistent.
func ApplyTimezone(cfg config.Config) error {
	loc, err := cfg.Location()
	if err != nil {
		return fmt.Errorf("load timezone: %w", err)
	}
	time.Local = loc
	return nil
}

// EnsureDirs makes sure the upload and database directories exist before any
// other code tries to write to them.
func EnsureDirs(cfg config.Config) error {
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return fmt.Errorf("create upload directory: %w", err)
	}
	if dbDir := filepath.Dir(cfg.DBPath); dbDir != "." {
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			return fmt.Errorf("create database directory: %w", err)
		}
	}
	return nil
}

// Open opens the sqlite database referenced by the config (or by an override
// path). Pass an empty override to use cfg.DBPath as-is.
func Open(cfg config.Config, override string) (*gorm.DB, error) {
	path := cfg.DBPath
	if override != "" {
		path = override
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database %q: %w", path, err)
	}
	return db, nil
}

// Migrate runs GORM AutoMigrate against every registered model.
func Migrate(db *gorm.DB) error {
	if err := dropLegacyRoomNoUniqueIndex(db); err != nil {
		return fmt.Errorf("relax room_no unique index: %w", err)
	}
	if err := db.AutoMigrate(AllModels()...); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

// dropLegacyRoomNoUniqueIndex removes the historical UNIQUE index on
// rooms.room_no. The model now uses a non-unique index, but GORM AutoMigrate
// does not downgrade an existing UNIQUE index, so we drop it explicitly. The
// non-unique index is re-created by AutoMigrate immediately after.
func dropLegacyRoomNoUniqueIndex(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.Room{}) {
		return nil
	}
	var count int64
	const probe = `SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'index'
		  AND tbl_name = 'rooms'
		  AND name = 'idx_rooms_room_no'
		  AND sql LIKE 'CREATE UNIQUE%'`
	if err := db.Raw(probe).Scan(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	return db.Exec("DROP INDEX IF EXISTS idx_rooms_room_no").Error
}

// Reset drops every managed table and then re-runs the migrations. Intended
// only for the seed command — the server entry point should never call this.
func Reset(db *gorm.DB) error {
	if err := db.Migrator().DropTable(AllModels()...); err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}
	return Migrate(db)
}
