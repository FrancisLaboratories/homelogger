package main

import (
	"encoding/json"
	"os"
	"testing"

	"time"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
)

// TestExportToJSON tests the export functionality — runs on the active test dialect.
func TestExportToJSON(t *testing.T) {
	db := database.TestDB(t)

	appliance := &models.Appliance{
		ApplianceName: "Test Fridge",
		Manufacturer:  "Samsung",
		ModelNumber:   "RF28R7201SR",
		SerialNumber:  "SN123456",
		YearPurchased: "2020",
		PurchasePrice: "1500",
		Location:      "Kitchen",
		Type:          "Refrigerator",
	}
	if err := db.Create(appliance).Error; err != nil {
		t.Fatalf("failed to create test appliance: %v", err)
	}

	task := &models.Task{
		Label:    "Replace water filter",
		Notes:    "Every 6 months",
		Checked:  false,
		Priority: "High",
		UserID:   "user1",
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	payload, err := database.ExportToJSON(db, db.Dialector.Name())
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if payload.Version != database.BackupVersion {
		t.Errorf("expected version %s, got %s", database.BackupVersion, payload.Version)
	}

	if payload.DatabaseType != db.Dialector.Name() {
		t.Errorf("expected databaseType %q, got %q", db.Dialector.Name(), payload.DatabaseType)
	}

	if len(payload.Entities.Appliances) != 1 {
		t.Errorf("expected 1 appliance, got %d", len(payload.Entities.Appliances))
	}

	if len(payload.Entities.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(payload.Entities.Tasks))
	}

	if payload.Entities.Appliances[0].ApplianceName != "Test Fridge" {
		t.Errorf("expected appliance name 'Test Fridge', got %s", payload.Entities.Appliances[0].ApplianceName)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON marshalling resulted in empty data")
	}
}

// TestImportFromJSON tests replace-all import semantics across all dialects.
func TestImportFromJSON(t *testing.T) {
	db := database.TestDB(t)

	preExisting := &models.Appliance{
		ApplianceName: "Old Fridge",
		Manufacturer:  "OldCo",
		ModelNumber:   "OLD123",
		SerialNumber:  "OLDSN",
		YearPurchased: "2010",
		PurchasePrice: "500",
		Location:      "Garage",
		Type:          "Refrigerator",
	}
	if err := db.Create(preExisting).Error; err != nil {
		t.Fatalf("failed to create pre-existing appliance: %v", err)
	}

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		ExportedAt:   time.Now().UTC(),
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{
					ApplianceName: "New Dishwasher",
					Manufacturer:  "NewCo",
					ModelNumber:   "NEW456",
					SerialNumber:  "NEWSN",
					YearPurchased: "2023",
					PurchasePrice: "1000",
					Location:      "Kitchen",
					Type:          "Dishwasher",
				},
				{
					ApplianceName: "Backup Fridge",
					Manufacturer:  "BackupCo",
					ModelNumber:   "BKP123",
					SerialNumber:  "BSNSN",
					YearPurchased: "2020",
					PurchasePrice: "800",
					Location:      "Kitchen",
					Type:          "Refrigerator",
				},
			},
		},
	}

	tempDir, err := os.MkdirTemp("", "testfiles-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	result, err := database.ImportFromJSON(db, payload, tempDir)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if result.Errors > 0 {
		t.Errorf("import returned errors: %s", result.ErrorMessage)
	}
	if result.Inserted != 2 {
		t.Errorf("expected 2 inserted records, got %d", result.Inserted)
	}

	var appliances []models.Appliance
	if err := db.Order("id asc").Find(&appliances).Error; err != nil {
		t.Fatalf("failed to query appliances: %v", err)
	}

	if len(appliances) != 2 {
		t.Fatalf("expected 2 appliances in DB after import, got %d", len(appliances))
	}

	for _, app := range appliances {
		if app.ApplianceName == "Old Fridge" {
			t.Error("pre-existing 'Old Fridge' should have been wiped by import")
		}
	}

	names := map[string]bool{}
	for _, app := range appliances {
		names[app.ApplianceName] = true
	}
	if !names["New Dishwasher"] {
		t.Error("'New Dishwasher' from payload not found in DB")
	}
	if !names["Backup Fridge"] {
		t.Error("'Backup Fridge' from payload not found in DB")
	}
}


