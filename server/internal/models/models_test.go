package models

import (
    "testing"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// openInMemory opens a unique in-memory sqlite DB for the test.
func openInMemory(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open db: %v", err)
    }
    return db
}

func TestApplianceGormCRUD(t *testing.T) {
    db := openInMemory(t)

    // Migrate only the Appliance model for this test
    if err := db.AutoMigrate(&Appliance{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    // Create
    a := &Appliance{
        ApplianceName: "TestFridge",
        Manufacturer:  "Acme",
        ModelNumber:   "T-1",
        SerialNumber:  "SN1",
        YearPurchased: "2022",
        PurchasePrice: "500",
        Location:      "Kitchen",
        Type:          "Appliance",
    }
    if err := db.Create(a).Error; err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    if a.ID == 0 {
        t.Fatalf("expected non-zero ID after create")
    }

    // Read
    var got Appliance
    if err := db.First(&got, a.ID).Error; err != nil {
        t.Fatalf("First failed: %v", err)
    }
    if got.ApplianceName != a.ApplianceName {
        t.Fatalf("name mismatch: %s != %s", got.ApplianceName, a.ApplianceName)
    }

    // Update
    got.Location = "Garage"
    if err := db.Save(&got).Error; err != nil {
        t.Fatalf("Save failed: %v", err)
    }
    var after Appliance
    if err := db.First(&after, a.ID).Error; err != nil {
        t.Fatalf("First after update failed: %v", err)
    }
    if after.Location != "Garage" {
        t.Fatalf("update didn't persist: %s", after.Location)
    }

    // Delete (hard delete)
    if err := db.Delete(&Appliance{}, a.ID).Error; err != nil {
        t.Fatalf("Delete failed: %v", err)
    }
    var count int64
    db.Model(&Appliance{}).Where("id = ?", a.ID).Count(&count)
    if count != 0 {
        t.Fatalf("expected row removed, count=%d", count)
    }
}