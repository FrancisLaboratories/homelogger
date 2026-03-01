package models

import (
    "testing"
)

func TestRepairGormCRUD(t *testing.T) {
    db := openInMemory(t)

    if err := db.AutoMigrate(&Appliance{}, &SavedFile{}, &Repair{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    a := &Appliance{ApplianceName: "RApp", Manufacturer: "Rfg", ModelNumber: "R-1", SerialNumber: "RS1", YearPurchased: "2019", PurchasePrice: "200", Location: "Basement", Type: "Appliance"}
    if err := db.Create(a).Error; err != nil {
        t.Fatalf("create appliance: %v", err)
    }

    r := &Repair{Description: "Fix leak", Date: "2026-02-28", Cost: 45.0, Notes: "fixed", SpaceType: "Basement", ReferenceType: "appliance", ApplianceID: &a.ID}
    if err := db.Create(r).Error; err != nil {
        t.Fatalf("create repair: %v", err)
    }
    if r.ID == 0 {
        t.Fatalf("expected repair ID after create")
    }

    var got Repair
    if err := db.First(&got, r.ID).Error; err != nil {
        t.Fatalf("first repair: %v", err)
    }
    if got.Description != r.Description {
        t.Fatalf("description mismatch: %s != %s", got.Description, r.Description)
    }

    // update
    got.Notes = "updated"
    if err := db.Save(&got).Error; err != nil {
        t.Fatalf("save repair: %v", err)
    }

    var after Repair
    if err := db.First(&after, r.ID).Error; err != nil {
        t.Fatalf("first after update: %v", err)
    }
    if after.Notes != "updated" {
        t.Fatalf("notes didn't update: %s", after.Notes)
    }

    // delete
    if err := db.Delete(&Repair{}, r.ID).Error; err != nil {
        t.Fatalf("delete repair: %v", err)
    }
}
