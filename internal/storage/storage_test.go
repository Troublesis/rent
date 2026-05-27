package storage

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMigrateDropsLegacyRoomNoUniqueIndex simulates a database upgraded from
// the old schema where room_no had a UNIQUE index. After Migrate runs, the
// unique constraint must be gone and duplicate room_no inserts must succeed.
func TestMigrateDropsLegacyRoomNoUniqueIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	legacy := []string{
		`CREATE TABLE rooms (
			id INTEGER PRIMARY KEY,
			room_no TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'vacant',
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE UNIQUE INDEX idx_rooms_room_no ON rooms (room_no)`,
	}
	for _, stmt := range legacy {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed legacy schema: %v", err)
		}
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var uniqueCount int64
	probe := `SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'index'
		  AND tbl_name = 'rooms'
		  AND name = 'idx_rooms_room_no'
		  AND sql LIKE 'CREATE UNIQUE%'`
	if err := db.Raw(probe).Scan(&uniqueCount).Error; err != nil {
		t.Fatalf("probe: %v", err)
	}
	if uniqueCount != 0 {
		t.Fatalf("legacy unique index still present, want 0 unique-index rows got %d", uniqueCount)
	}

	if err := db.Exec(`INSERT INTO rooms (room_no, title) VALUES ('A101', 'one')`).Error; err != nil {
		t.Fatalf("insert first: %v", err)
	}
	if err := db.Exec(`INSERT INTO rooms (room_no, title) VALUES ('A101', 'two')`).Error; err != nil {
		t.Fatalf("duplicate room_no still rejected: %v", err)
	}
}
