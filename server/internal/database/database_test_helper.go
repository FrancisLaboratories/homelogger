package database

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testDB creates a unique in-memory sqlite DB per test (avoids shared cache collisions), migrates models, and returns the DB.
func testDB(t *testing.T) *gorm.DB {
    t.Helper()

    // Use a unique DSN per test to avoid sharing state between tests
    dsn := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
    db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open in-memory db: %v", err)
    }

    if err := MigrateGorm(db); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }

    return db
}
