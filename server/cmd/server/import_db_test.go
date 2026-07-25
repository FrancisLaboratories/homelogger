package main

import (
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestImportFromJSON_AllEntities(t *testing.T) {
	db := openTestDB(t)

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "Fridge"},
				{ApplianceName: "Washer"},
			},
			Todos: []models.Todo{
				{UserID: "user1"},
				{UserID: "user2"},
			},
			Maintenance: []models.Maintenance{
				{Description: "Filter change", Date: "2026-01-15"},
			},
			Repairs: []models.Repair{
				{Description: "Fix leak", Date: "2026-03-01"},
			},
			SavedFiles: []models.SavedFile{
				{Path: "./data/uploads/1", OriginalName: "doc.pdf", UserID: "u1"},
			},
			Notes: []models.Note{
				{Title: "Note 1", Body: "Body 1"},
			},
			Tasks: []models.Task{
				{Label: "Task 1", UserID: "u1"},
				{Label: "Task 2", UserID: "u2"},
			},
		},
	}

	result, err := database.ImportFromJSON(db, payload, "")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	expected := 10 // 2 + 2 + 1 + 1 + 1 + 1 + 2
	if result.Inserted != expected {
		t.Errorf("expected %d inserted records, got %d", expected, result.Inserted)
	}

	var appliances []models.Appliance
	if err := db.Find(&appliances).Error; err != nil {
		t.Fatalf("query appliances: %v", err)
	}
	if len(appliances) != 2 {
		t.Errorf("expected 2 appliances, got %d", len(appliances))
	}

	var todos []models.Todo
	if err := db.Find(&todos).Error; err != nil {
		t.Fatalf("query todos: %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}

	var maintenance []models.Maintenance
	if err := db.Find(&maintenance).Error; err != nil {
		t.Fatalf("query maintenance: %v", err)
	}
	if len(maintenance) != 1 {
		t.Errorf("expected 1 maintenance, got %d", len(maintenance))
	}

	var repairs []models.Repair
	if err := db.Find(&repairs).Error; err != nil {
		t.Fatalf("query repairs: %v", err)
	}
	if len(repairs) != 1 {
		t.Errorf("expected 1 repair, got %d", len(repairs))
	}

	var savedFiles []models.SavedFile
	if err := db.Find(&savedFiles).Error; err != nil {
		t.Fatalf("query savedFiles: %v", err)
	}
	if len(savedFiles) != 1 {
		t.Errorf("expected 1 savedFile, got %d", len(savedFiles))
	}

	var notes []models.Note
	if err := db.Find(&notes).Error; err != nil {
		t.Fatalf("query notes: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}

	var tasks []models.Task
	if err := db.Find(&tasks).Error; err != nil {
		t.Fatalf("query tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestImportFromJSON_EmptyPayload(t *testing.T) {
	db := openTestDB(t)

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
	}

	result, err := database.ImportFromJSON(db, payload, "")
	if err != nil {
		t.Fatalf("expected success with empty payload, got %v", err)
	}
	if result.Inserted != 0 {
		t.Errorf("expected 0 inserted, got %d", result.Inserted)
	}

	// Verify all tables exist and are empty
	var count int64
	db.Model(&models.Appliance{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 appliances, got %d", count)
	}
}

func TestImportFromJSON_UploadValidationFails(t *testing.T) {
	db := openTestDB(t)

	tempDir := t.TempDir()
	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			SavedFiles: []models.SavedFile{
				{Path: "./data/uploads/missing-file", OriginalName: "doc.pdf", UserID: "u1"},
			},
		},
	}

	// The validation should fail because missing-file doesn't exist in tempDir
	_, err := database.ImportFromJSON(db, payload, tempDir)
	if err == nil {
		t.Fatal("expected error for missing upload file, got nil")
	}
}

func TestImportFromJSON_WipesExistingData(t *testing.T) {
	db := openTestDB(t)

	// Insert pre-existing data
	if err := db.Create(&models.Appliance{ApplianceName: "Old Fridge"}).Error; err != nil {
		t.Fatalf("insert pre-existing: %v", err)
	}

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "New Fridge"},
			},
		},
	}

	result, err := database.ImportFromJSON(db, payload, "")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result.Inserted != 1 {
		t.Errorf("expected 1 inserted, got %d", result.Inserted)
	}

	var count int64
	db.Model(&models.Appliance{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 appliance after import, got %d", count)
	}

	var appliances []models.Appliance
	db.Find(&appliances)
	if appliances[0].ApplianceName != "New Fridge" {
		t.Errorf("expected 'New Fridge', got %q", appliances[0].ApplianceName)
	}
}


