package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func validMinimalPayload() *models.BackupPayload {
	return &models.BackupPayload{
		Version:      BackupVersion,
		DatabaseType: "sqlite",
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "Test Fridge"},
			},
			Maintenance: []models.Maintenance{
				{Description: "Filter change", Date: "2026-01-15"},
			},
			Repairs: []models.Repair{
				{Description: "Fix leak", Date: "2026-03-01"},
			},
			SavedFiles: []models.SavedFile{
				{Path: "./data/uploads/1", OriginalName: "doc.pdf", UserID: "user1"},
			},
			Todos: []models.Todo{
				{UserID: "user1"},
			},
		},
	}
}

func TestValidatePayload(t *testing.T) {
	t.Run("valid minimal payload", func(t *testing.T) {
		err := validatePayload(validMinimalPayload())
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("nil payload", func(t *testing.T) {
		err := validatePayload(nil)
		if err == nil {
			t.Fatal("expected error for nil payload")
		}
	})

	t.Run("empty version", func(t *testing.T) {
		p := validMinimalPayload()
		p.Version = ""
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for empty version")
		}
	})

	t.Run("empty database type", func(t *testing.T) {
		p := validMinimalPayload()
		p.DatabaseType = ""
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for empty database type")
		}
	})

	t.Run("appliance missing name", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Appliances = append(p.Entities.Appliances, models.Appliance{})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for appliance with empty name")
		}
	})

	t.Run("maintenance missing description", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Maintenance = append(p.Entities.Maintenance, models.Maintenance{Date: "2026-01-01"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for maintenance with empty description")
		}
	})

	t.Run("maintenance missing date", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Maintenance = append(p.Entities.Maintenance, models.Maintenance{Description: "Test"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for maintenance with empty date")
		}
	})

	t.Run("repair missing description", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Repairs = append(p.Entities.Repairs, models.Repair{Date: "2026-01-01"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for repair with empty description")
		}
	})

	t.Run("repair missing date", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Repairs = append(p.Entities.Repairs, models.Repair{Description: "Test"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for repair with empty date")
		}
	})

	t.Run("savedFile missing path", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.SavedFiles = append(p.Entities.SavedFiles, models.SavedFile{OriginalName: "f", UserID: "u"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for savedFile with empty path")
		}
	})

	t.Run("savedFile missing originalName", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.SavedFiles = append(p.Entities.SavedFiles, models.SavedFile{Path: "./p", UserID: "u"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for savedFile with empty originalName")
		}
	})

	t.Run("savedFile missing userid", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.SavedFiles = append(p.Entities.SavedFiles, models.SavedFile{Path: "./p", OriginalName: "f"})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for savedFile with empty userid")
		}
	})

	t.Run("todo missing userid", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Todos = append(p.Entities.Todos, models.Todo{})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for todo with empty userid")
		}
	})

	t.Run("empty entities sections should pass", func(t *testing.T) {
		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
		}
		err := validatePayload(p)
		if err != nil {
			t.Fatalf("expected nil for empty entities, got %v", err)
		}
	})

	t.Run("all entity types with empty required fields in one payload", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Appliances = append(p.Entities.Appliances, models.Appliance{})
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestValidatePayload_Duplicates(t *testing.T) {
	t.Run("duplicate appliance IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Appliances = []models.Appliance{
			{ApplianceName: "A", ID: 1},
			{ApplianceName: "B", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate appliance IDs")
		}
	})

	t.Run("duplicate todo IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Todos = []models.Todo{
			{UserID: "u", ID: 1},
			{UserID: "v", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate todo IDs")
		}
	})

	t.Run("duplicate maintenance IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Maintenance = []models.Maintenance{
			{Description: "A", Date: "2026-01-01", ID: 1},
			{Description: "B", Date: "2026-01-02", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate maintenance IDs")
		}
	})

	t.Run("duplicate repair IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Repairs = []models.Repair{
			{Description: "A", Date: "2026-01-01", ID: 1},
			{Description: "B", Date: "2026-01-02", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate repair IDs")
		}
	})

	t.Run("duplicate savedFile IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.SavedFiles = []models.SavedFile{
			{Path: "./a", OriginalName: "a", UserID: "u", ID: 1},
			{Path: "./b", OriginalName: "b", UserID: "v", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate savedFile IDs")
		}
	})

	t.Run("duplicate note IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Notes = []models.Note{
			{Body: "note1", ID: 1},
			{Body: "note2", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate note IDs")
		}
	})

	t.Run("duplicate task IDs", func(t *testing.T) {
		p := validMinimalPayload()
		p.Entities.Tasks = []models.Task{
			{Label: "task1", ID: 1},
			{Label: "task2", ID: 1},
		}
		err := validatePayload(p)
		if err == nil {
			t.Fatal("expected error for duplicate task IDs")
		}
	})
}

func TestValidateUploads(t *testing.T) {
	t.Run("empty uploadsDir skips validation", func(t *testing.T) {
		p := validMinimalPayload()
		err := validateUploads(p, "")
		if err != nil {
			t.Fatalf("expected nil for empty uploadsDir, got %v", err)
		}
	})

	t.Run("all files exist", func(t *testing.T) {
		dir := t.TempDir()
		rel := "1"
		os.MkdirAll(filepath.Join(dir, filepath.Dir(rel)), 0755)
		os.WriteFile(filepath.Join(dir, rel), []byte("data"), 0644)

		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{Path: "./data/uploads/1", OriginalName: "doc.pdf", UserID: "u"},
				},
			},
		}
		err := validateUploads(p, dir)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		dir := t.TempDir()
		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{Path: "./data/uploads/missing-file", OriginalName: "doc.pdf", UserID: "u"},
				},
			},
		}
		err := validateUploads(p, dir)
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("uploadsDir does not exist on disk", func(t *testing.T) {
		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{Path: "./data/uploads/1", OriginalName: "doc.pdf", UserID: "u"},
				},
			},
		}
		err := validateUploads(p, "./nonexistent-dir")
		if err == nil {
			t.Fatal("expected error for non-existent uploads dir")
		}
	})

	t.Run("no saved files should pass", func(t *testing.T) {
		dir := t.TempDir()
		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
		}
		err := validateUploads(p, dir)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("savedFile with empty path is skipped", func(t *testing.T) {
		dir := t.TempDir()
		p := &models.BackupPayload{
			Version:      BackupVersion,
			DatabaseType: "sqlite",
			Entities: models.Entities{
				SavedFiles: []models.SavedFile{
					{Path: "", OriginalName: "doc.pdf", UserID: "u"},
				},
			},
		}
		err := validateUploads(p, dir)
		if err != nil {
			t.Fatalf("expected nil for savedFile with empty path, got %v", err)
		}
	})
}
