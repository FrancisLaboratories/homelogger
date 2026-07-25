package database

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

func createLegacyDB(t *testing.T, tables map[string]bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	if tables["appliances"] {
		db.Exec(`CREATE TABLE appliances (id INTEGER PRIMARY KEY, appliance_name TEXT, manufacturer TEXT, model_number TEXT, serial_number TEXT, year_purchased TEXT, purchase_price TEXT, location TEXT, type TEXT, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO appliances (id, appliance_name, manufacturer) VALUES (1, 'Old Fridge', 'Acme')`)
	}

	if tables["todos"] {
		db.Exec(`CREATE TABLE todos (id INTEGER PRIMARY KEY, label TEXT, checked BOOLEAN, userid TEXT, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO todos (id, label, checked, userid) VALUES (1, 'Buy filter', 0, 'u1')`)
	}

	if tables["tasks"] {
		db.Exec(`CREATE TABLE tasks (id INTEGER PRIMARY KEY, label TEXT, userid TEXT, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO tasks (id, label, userid) VALUES (1, 'Replace filter', 'u1')`)
	}

	if tables["maintenances"] {
		db.Exec(`CREATE TABLE maintenances (id INTEGER PRIMARY KEY, description TEXT, date TEXT, cost REAL, notes TEXT, space_type TEXT, reference_type TEXT, appliance_id INTEGER, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO maintenances (id, description, date) VALUES (1, 'Filter change', '2026-01-15')`)
	}

	if tables["repairs"] {
		db.Exec(`CREATE TABLE repairs (id INTEGER PRIMARY KEY, description TEXT, date TEXT, cost REAL, notes TEXT, space_type TEXT, reference_type TEXT, appliance_id INTEGER, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO repairs (id, description, date) VALUES (1, 'Fix leak', '2026-03-01')`)
	}

	if tables["saved_files"] {
		db.Exec(`CREATE TABLE saved_files (id INTEGER PRIMARY KEY, path TEXT, original_name TEXT, type TEXT, userid TEXT, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO saved_files (id, path, original_name, userid) VALUES (1, './data/uploads/1', 'doc.pdf', 'u1')`)
	}

	if tables["notes"] {
		db.Exec(`CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT, body TEXT, "createdAt" DATETIME, "updatedAt" DATETIME, "deletedAt" DATETIME)`)
		db.Exec(`INSERT INTO notes (id, title, body) VALUES (1, 'Note 1', 'Body 1')`)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	return path
}

func TestConvertLegacyDB(t *testing.T) {
	t.Run("all tables present", func(t *testing.T) {
		path := createLegacyDB(t, map[string]bool{
			"appliances":  true,
			"todos":       true,
			"tasks":       true,
			"maintenances": true,
			"repairs":     true,
			"saved_files": true,
			"notes":       true,
		})

		payload, err := ConvertLegacyDB(path)
		if err != nil {
			t.Fatalf("ConvertLegacyDB failed: %v", err)
		}

		if payload.Version != BackupVersion {
			t.Errorf("expected version %q, got %q", BackupVersion, payload.Version)
		}
		if payload.DatabaseType != "sqlite" {
			t.Errorf("expected databaseType sqlite, got %q", payload.DatabaseType)
		}

		if len(payload.Entities.Appliances) != 1 {
			t.Errorf("expected 1 appliance, got %d", len(payload.Entities.Appliances))
		}
		if len(payload.Entities.Todos) != 1 {
			t.Errorf("expected 1 todo, got %d", len(payload.Entities.Todos))
		}
		if len(payload.Entities.Tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(payload.Entities.Tasks))
		}
		if len(payload.Entities.Maintenance) != 1 {
			t.Errorf("expected 1 maintenance, got %d", len(payload.Entities.Maintenance))
		}
		if len(payload.Entities.Repairs) != 1 {
			t.Errorf("expected 1 repair, got %d", len(payload.Entities.Repairs))
		}
		if len(payload.Entities.SavedFiles) != 1 {
			t.Errorf("expected 1 savedFile, got %d", len(payload.Entities.SavedFiles))
		}
		if len(payload.Entities.Notes) != 1 {
			t.Errorf("expected 1 note, got %d", len(payload.Entities.Notes))
		}
	})

	t.Run("missing tables skipped gracefully", func(t *testing.T) {
		path := createLegacyDB(t, map[string]bool{
			"appliances": true,
		})

		payload, err := ConvertLegacyDB(path)
		if err != nil {
			t.Fatalf("ConvertLegacyDB failed: %v", err)
		}

		if len(payload.Entities.Appliances) != 1 {
			t.Errorf("expected 1 appliance, got %d", len(payload.Entities.Appliances))
		}
		if len(payload.Entities.Tasks) != 0 {
			t.Errorf("expected 0 tasks (no table), got %d", len(payload.Entities.Tasks))
		}
		if len(payload.Entities.Notes) != 0 {
			t.Errorf("expected 0 notes (no table), got %d", len(payload.Entities.Notes))
		}
	})
}

func TestSanitizeFKs(t *testing.T) {
	t.Run("orphaned appliance FK on maintenance is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Maintenance: []models.Maintenance{
					{ApplianceID: uintPtr(99)}, // no appliance with ID 99
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Maintenance[0].ApplianceID != nil {
			t.Error("expected ApplianceID to be nil")
		}
	})

	t.Run("valid appliance FK on maintenance is preserved", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Appliances: []models.Appliance{
					{ApplianceName: "Fridge", ID: 1},
				},
				Maintenance: []models.Maintenance{
					{ApplianceID: uintPtr(1)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Maintenance[0].ApplianceID == nil || *payload.Entities.Maintenance[0].ApplianceID != 1 {
			t.Error("expected ApplianceID to remain 1")
		}
	})

	t.Run("orphaned attachment FK on repair is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Repairs: []models.Repair{
					{AttachmentID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Repairs[0].AttachmentID != nil {
			t.Error("expected AttachmentID to be nil")
		}
	})

	t.Run("orphaned attachment FK on maintenance is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Maintenance: []models.Maintenance{
					{AttachmentID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Maintenance[0].AttachmentID != nil {
			t.Error("expected AttachmentID to be nil")
		}
	})

	t.Run("orphaned appliance FK on savedFile is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{ApplianceID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.SavedFiles[0].ApplianceID != nil {
			t.Error("expected ApplianceID to be nil")
		}
	})

	t.Run("orphaned maintenance FK on savedFile is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{MaintenanceID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.SavedFiles[0].MaintenanceID != nil {
			t.Error("expected MaintenanceID to be nil")
		}
	})

	t.Run("orphaned repair FK on savedFile is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{RepairID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.SavedFiles[0].RepairID != nil {
			t.Error("expected RepairID to be nil")
		}
	})

	t.Run("orphaned appliance FK on note is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Notes: []models.Note{
					{ApplianceID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Notes[0].ApplianceID != nil {
			t.Error("expected ApplianceID to be nil")
		}
	})

	t.Run("orphaned appliance FK on task is nilled", func(t *testing.T) {
		payload := &models.BackupPayload{
			Entities: models.Entities{
				Tasks: []models.Task{
					{ApplianceID: uintPtr(99)},
				},
			},
		}
		SanitizeFKs(payload)
		if payload.Entities.Tasks[0].ApplianceID != nil {
			t.Error("expected ApplianceID to be nil")
		}
	})
}


