package demo

import (
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openInMemoryDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file:mem_demo_demo?mode=memory&cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open in-memory db: %v", err)
    }
    if err := database.MigrateGorm(db); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    return db
}

func TestSeedLoadsSampleData(t *testing.T) {
    db := openInMemoryDB(t)

    // sample_data.json lives in the same package directory as this test,
    // so reference it relative to the package (working dir during `go test`).
    demoPath := "sample_data.json"
    if err := Seed(db, demoPath); err != nil {
        t.Fatalf("Seed returned error: %v", err)
    }

    // appliances
    apps, err := database.GetAppliances(db)
    if err != nil {
        t.Fatalf("GetAppliances error: %v", err)
    }
    if len(apps) != 2 {
        t.Fatalf("expected 2 appliances from demo data, got %d", len(apps))
    }

    // todos for first appliance should include a known label
    aid := apps[0].ID
    todos, err := database.GetTodos(db, aid, "")
    if err != nil {
        t.Fatalf("GetTodos error: %v", err)
    }
    if len(todos) == 0 {
        t.Fatalf("expected todos for appliance %d, got 0", aid)
    }

    // notes for first appliance
    notes, err := database.GetNotes(db, aid, "")
    if err != nil {
        t.Fatalf("GetNotes error: %v", err)
    }
    if len(notes) == 0 {
        t.Fatalf("expected notes for appliance %d, got 0", aid)
    }

    // maintenances and repairs should exist for appliance 0
    maint, err := database.GetMaintenances(db, aid, "Appliance", "")
    if err != nil {
        t.Fatalf("GetMaintenances error: %v", err)
    }
    if len(maint) == 0 {
        t.Fatalf("expected maintenances for appliance %d, got 0", aid)
    }

    rep, err := database.GetRepairs(db, aid, "Appliance", "")
    if err != nil {
        t.Fatalf("GetRepairs error: %v", err)
    }
    if len(rep) == 0 {
        t.Fatalf("expected repairs for appliance %d, got 0", aid)
    }
}

func TestSeedMissingFile(t *testing.T) {
    db := openInMemoryDB(t)
    if err := Seed(db, "nonexistent_demo_file.json"); err == nil {
        t.Fatalf("expected error when demo file missing")
    }
}
