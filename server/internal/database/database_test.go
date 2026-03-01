package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// setupTestDB prepares a temporary sqlite DB and runs migrations.
func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")
    if err := os.Setenv("DEMO_DB_PATH", dbPath); err != nil {
        t.Fatalf("failed to set DEMO_DB_PATH: %v", err)
    }

    db, err := ConnectGorm()
    if err != nil {
        t.Fatalf("ConnectGorm error: %v", err)
    }

    if err := MigrateGorm(db); err != nil {
        t.Fatalf("MigrateGorm error: %v", err)
    }

    return db
}

func TestConnectAndMigrate(t *testing.T) {
    db := setupTestDB(t)
    sqlDB, err := db.DB()
    if err != nil {
        t.Fatalf("db.DB() error: %v", err)
    }
    defer sqlDB.Close()
}

func TestApplianceCRUD(t *testing.T) {
    db := setupTestDB(t)

    ap := &models.Appliance{
        ApplianceName: "Dishwasher",
        Manufacturer:  "Acme",
        ModelNumber:   "D-100",
        SerialNumber:  "SN123",
        YearPurchased: "2020",
        PurchasePrice: "200",
        Location:      "Kitchen",
        Type:          "Appliance",
    }

    // Add
    added, err := AddAppliance(db, ap)
    if err != nil {
        t.Fatalf("AddAppliance error: %v", err)
    }
    if added.ID == 0 {
        t.Fatalf("expected non-zero ID after add")
    }

    // GetAll
    list, err := GetAppliances(db)
    if err != nil {
        t.Fatalf("GetAppliances error: %v", err)
    }
    if len(list) != 1 {
        t.Fatalf("expected 1 appliance, got %d", len(list))
    }

    // GetByID
    fetched, err := GetAppliance(db, added.ID)
    if err != nil {
        t.Fatalf("GetAppliance error: %v", err)
    }
    if fetched.ApplianceName != ap.ApplianceName {
        t.Fatalf("unexpected name: %s", fetched.ApplianceName)
    }

    // Update
    fetched.Location = "Basement"
    updated, err := UpdateAppliance(db, fetched)
    if err != nil {
        t.Fatalf("UpdateAppliance error: %v", err)
    }
    if updated.Location != "Basement" {
        t.Fatalf("expected updated location, got %s", updated.Location)
    }

    // Delete
    if err := DeleteAppliance(db, updated.ID); err != nil {
        t.Fatalf("DeleteAppliance error: %v", err)
    }

    // Ensure deleted
    if _, err := GetAppliance(db, updated.ID); err == nil {
        t.Fatalf("expected error fetching deleted appliance")
    }
}

func TestFileAttachAndDelete(t *testing.T) {
    db := setupTestDB(t)

    // Create an appliance to attach files to
    ap := &models.Appliance{ApplianceName: "Fridge", Manufacturer: "Acme", ModelNumber: "F1", SerialNumber: "S1", YearPurchased: "2021", PurchasePrice: "300", Location: "Kitchen", Type: "Appliance"}
    added, err := AddAppliance(db, ap)
    if err != nil {
        t.Fatalf("AddAppliance error: %v", err)
    }

    // create a real file on disk
    tmp := t.TempDir()
    fp := filepath.Join(tmp, "sample.txt")
    if err := os.WriteFile(fp, []byte("hello"), 0644); err != nil {
        t.Fatalf("failed writing sample file: %v", err)
    }

    // Upload file record pointing to that path
    sf := &models.SavedFile{Path: fp, OriginalName: "sample.txt", UserID: "u1"}
    // Use UploadFile to create entry
    uploaded, err := UploadFile(db, sf)
    if err != nil {
        t.Fatalf("UploadFile error: %v", err)
    }

    // Attach to appliance
    if err := AttachFileToAppliance(db, uploaded.ID, added.ID); err != nil {
        t.Fatalf("AttachFileToAppliance error: %v", err)
    }

    // Ensure file is listed by appliance
    files, err := GetFilesByAppliance(db, added.ID)
    if err != nil {
        t.Fatalf("GetFilesByAppliance error: %v", err)
    }
    if len(files) != 1 {
        t.Fatalf("expected 1 file, got %d", len(files))
    }

    // Delete files by appliance
    if err := DeleteFilesByAppliance(db, added.ID); err != nil {
        t.Fatalf("DeleteFilesByAppliance error: %v", err)
    }

    // file should be removed from disk
    if _, err := os.Stat(fp); !os.IsNotExist(err) {
        t.Fatalf("expected file to be removed from disk")
    }

    // DB should have zero files for appliance
    filesAfter, err := GetFilesByAppliance(db, added.ID)
    if err != nil {
        t.Fatalf("GetFilesByAppliance after delete error: %v", err)
    }
    if len(filesAfter) != 0 {
        t.Fatalf("expected 0 files after delete, got %d", len(filesAfter))
    }
}

