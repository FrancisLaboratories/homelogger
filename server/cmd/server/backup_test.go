package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"time"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestExportToJSON tests the export functionality
func TestExportToJSON(t *testing.T) {
	// Setup: Create test database with some data
	db, err := setupTestDB(database.DialectSQLite)
	if err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db, database.DialectSQLite)

	// Insert test data
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

	// Export to JSON
	payload, err := database.ExportToJSON(db, "sqlite")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Verify payload structure
	if payload.Version != database.BackupVersion {
		t.Errorf("expected version %s, got %s", database.BackupVersion, payload.Version)
	}

	if payload.DatabaseType != "sqlite" {
		t.Errorf("expected databaseType 'sqlite', got %s", payload.DatabaseType)
	}

	if len(payload.Entities.Appliances) != 1 {
		t.Errorf("expected 1 appliance, got %d", len(payload.Entities.Appliances))
	}

	if len(payload.Entities.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(payload.Entities.Tasks))
	}

	// Verify appliance data
	if payload.Entities.Appliances[0].ApplianceName != "Test Fridge" {
		t.Errorf("expected appliance name 'Test Fridge', got %s", payload.Entities.Appliances[0].ApplianceName)
	}

	// Verify JSON marshalling
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON marshalling resulted in empty data")
	}
}

// TestExportToJSONPostgres tests the export functionality for PostgreSQL
func TestExportToJSONPostgres(t *testing.T) {
	t.Skip("PostgreSQL test setup not implemented yet")

	// Setup: Create test database with some data
	db, err := setupTestDB(database.DialectPostgres)
	if err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db, database.DialectPostgres)

	// Insert test data
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

	// Export to JSON
	payload, err := database.ExportToJSON(db, database.DialectPostgres)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Verify payload structure
	if payload.Version != database.BackupVersion {
		t.Errorf("expected version %s, got %s", database.BackupVersion, payload.Version)
	}

	if payload.DatabaseType != database.DialectPostgres {
		t.Errorf("expected databaseType '%s', got %s", database.DialectPostgres, payload.DatabaseType)
	}

	if len(payload.Entities.Appliances) != 1 {
		t.Errorf("expected 1 appliance, got %d", len(payload.Entities.Appliances))
	}

	if len(payload.Entities.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(payload.Entities.Tasks))
	}

	// Verify appliance data
	if payload.Entities.Appliances[0].ApplianceName != "Test Fridge" {
		t.Errorf("expected appliance name 'Test Fridge', got %s", payload.Entities.Appliances[0].ApplianceName)
	}

	// Verify JSON marshalling
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON marshalling resulted in empty data")
	}
}


// TestImportFromJSON tests replace-all import semantics:
// pre-existing DB data is wiped; payload records are inserted fresh.
func TestImportFromJSON(t *testing.T) {
	db, err := setupTestDB(database.DialectSQLite)
	if err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db, database.DialectSQLite)

	// Seed a record that should be gone after import
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

	// Payload contains two appliances (neither is "Old Fridge")
	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		ExportedAt:   time.Now().UTC(),
		DatabaseType: "sqlite",
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
	// Replace semantics: both payload appliances inserted, pre-existing wiped
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

	// Pre-existing "Old Fridge" must be gone
	for _, app := range appliances {
		if app.ApplianceName == "Old Fridge" {
			t.Error("pre-existing 'Old Fridge' should have been wiped by import")
		}
	}

	// Both payload appliances must be present
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

// TestImportMergeConflict tests that newer UpdatedAt wins
// TODO: Fix test fixtures
func TestImportMergeConflict(t *testing.T) {
	t.Skip("test fixtures need refinement")
	/* Placeholder for merge conflict resolution */
}

// TestImportMissingFK tests that records with missing foreign keys are skipped
// TODO: Fix test fixtures
func TestImportMissingFK(t *testing.T) {
	t.Skip("test fixtures need refinement")
	/* Placeholder for FK validation */
}

// TestImportSchemaMismatch tests schema version validation
// TODO: Fix test fixtures
func TestImportSchemaMismatch(t *testing.T) {
	t.Skip("test fixtures need refinement")
	/* Placeholder for schema validation */
}

func setupTestDB(dbType string) (*gorm.DB, error) {
	if dbType == database.DialectPostgres {
		return nil, fmt.Errorf("PostgreSQL test setup not implemented yet")
	}

	// Create temporary database
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	// Open database
	db, err := gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate models
	if err := database.MigrateGorm(db); err != nil {
		return nil, err
	}

	return db, nil
}

func cleanupTestDB(db *gorm.DB, dbType string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
