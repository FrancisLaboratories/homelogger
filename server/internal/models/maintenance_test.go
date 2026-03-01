package models

import (
    "testing"
)

func TestMaintenanceGormCRUD(t *testing.T) {
    db := openInMemory(t)

    if err := db.AutoMigrate(&Appliance{}, &SavedFile{}, &Maintenance{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    // create an appliance to reference (optional)
    a := &Appliance{ApplianceName: "MApp", Manufacturer: "Mfg", ModelNumber: "M-1", SerialNumber: "S1", YearPurchased: "2020", PurchasePrice: "100", Location: "Attic", Type: "Appliance"}
    if err := db.Create(a).Error; err != nil {
        t.Fatalf("create appliance: %v", err)
    }

    m := &Maintenance{Description: "Check filters", Date: "2026-02-28", Cost: 12.5, Notes: "note", SpaceType: "Attic", ReferenceType: "appliance", ApplianceID: &a.ID}
    if err := db.Create(m).Error; err != nil {
        t.Fatalf("create maintenance: %v", err)
    }
    if m.ID == 0 {
        t.Fatalf("expected maintenance ID after create")
    }

    var got Maintenance
    if err := db.First(&got, m.ID).Error; err != nil {
        t.Fatalf("first maintenance: %v", err)
    }
    if got.Description != m.Description {
        t.Fatalf("description mismatch: %s != %s", got.Description, m.Description)
    }

    // update
    got.Cost = 99.9
    if err := db.Save(&got).Error; err != nil {
        t.Fatalf("save maintenance: %v", err)
    }

    var after Maintenance
    if err := db.First(&after, m.ID).Error; err != nil {
        t.Fatalf("first after update: %v", err)
    }
    if after.Cost != 99.9 {
        t.Fatalf("cost didn't update: %v", after.Cost)
    }

    // delete
    if err := db.Delete(&Maintenance{}, m.ID).Error; err != nil {
        t.Fatalf("delete maintenance: %v", err)
    }
}
