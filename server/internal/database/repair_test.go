package database

import (
    "os"
    "testing"

    "github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestAddGetDeleteRepair(t *testing.T) {
    db := testDB(t)

    // create appliance
    a := &models.Appliance{ApplianceName: "A", Manufacturer: "M", ModelNumber: "X", SerialNumber: "S", YearPurchased: "2020", PurchasePrice: "1", Location: "L", Type: "T"}
    if _, err := AddAppliance(db, a); err != nil {
        t.Fatalf("AddAppliance failed: %v", err)
    }

    r := &models.Repair{Description: "r1", ReferenceType: "Appliance", SpaceType: "", Date: "2026-01-02", ApplianceID: &a.ID}
    added, err := AddRepair(db, r)
    if err != nil {
        t.Fatalf("AddRepair failed: %v", err)
    }

    got, err := GetRepair(db, added.ID)
    if err != nil {
        t.Fatalf("GetRepair failed: %v", err)
    }
    if got.ID != added.ID {
        t.Fatalf("repair ID mismatch: %d vs %d", got.ID, added.ID)
    }

    // attach a file
    tmp, err := os.CreateTemp(os.TempDir(), "hl-r-*")
    if err != nil {
        t.Fatalf("failed create tmp: %v", err)
    }
    path := tmp.Name()
    tmp.Close()
    f := &models.SavedFile{Path: path, OriginalName: "rf", UserID: "u", RepairID: &added.ID}
    if _, err := UploadFile(db, f); err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    if err := DeleteRepair(db, added.ID); err != nil {
        t.Fatalf("DeleteRepair failed: %v", err)
    }

    if _, err := os.Stat(path); !os.IsNotExist(err) {
        t.Fatalf("expected repair file removed, stat error: %v", err)
    }
}

func TestGetRepairsSpaceFilter(t *testing.T) {
    db := testDB(t)

    r := &models.Repair{Description: "s1", ReferenceType: "Space", SpaceType: "Basement", Date: "2026-02-02"}
    if _, err := AddRepair(db, r); err != nil {
        t.Fatalf("AddRepair failed: %v", err)
    }

    res, err := GetRepairs(db, 0, "Space", "Basement")
    if err != nil {
        t.Fatalf("GetRepairs failed: %v", err)
    }
    if len(res) != 1 {
        t.Fatalf("expected 1 repair for space, got %d", len(res))
    }
}
