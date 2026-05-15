package database

import (
	"os"
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestAddGetDeleteMaintenance(t *testing.T) {
    db := testDB(t)

    // create appliance for non-space maintenance
    a := &models.Appliance{ApplianceName: "A", Manufacturer: "M", ModelNumber: "X", SerialNumber: "S", YearPurchased: "2020", PurchasePrice: "1", Location: "L", Type: "T"}
    if _, err := AddAppliance(db, a); err != nil {
        t.Fatalf("AddAppliance failed: %v", err)
    }

    m := &models.Maintenance{Description: "m1", ReferenceType: "Appliance", SpaceType: "", Date: "2026-01-01", ApplianceID: &a.ID}
    added, err := AddMaintenance(db, m)
    if err != nil {
        t.Fatalf("AddMaintenance failed: %v", err)
    }

    got, err := GetMaintenance(db, added.ID)
    if err != nil {
        t.Fatalf("GetMaintenance failed: %v", err)
    }
    if got.ID != added.ID {
        t.Fatalf("maintenance ID mismatch: %d vs %d", got.ID, added.ID)
    }

    // create a temp file and attach to maintenance
    tmp, err := os.CreateTemp(os.TempDir(), "hl-m-*")
    if err != nil {
        t.Fatalf("failed create tmp: %v", err)
    }
    path := tmp.Name()
    tmp.Close()
    f := &models.SavedFile{Path: path, OriginalName: "mf", UserID: "u", MaintenanceID: &added.ID}
    if _, err := UploadFile(db, f); err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // Delete maintenance should remove files and DB row
    if err := DeleteMaintenance(db, added.ID); err != nil {
        t.Fatalf("DeleteMaintenance failed: %v", err)
    }

    if _, err := os.Stat(path); !os.IsNotExist(err) {
        t.Fatalf("expected maintenance file removed, stat error: %v", err)
    }
}

func TestGetMaintenancesSpaceFilter(t *testing.T) {
    db := testDB(t)

    m := &models.Maintenance{Description: "s1", ReferenceType: "Space", SpaceType: "Yard", Date: "2026-02-01"}
    if _, err := AddMaintenance(db, m); err != nil {
        t.Fatalf("AddMaintenance failed: %v", err)
    }

    res, err := GetMaintenances(db, 0, "Space", "Yard")
    if err != nil {
        t.Fatalf("GetMaintenances failed: %v", err)
    }
    if len(res) != 1 {
        t.Fatalf("expected 1 maintenance for space, got %d", len(res))
    }
}

func TestUpdateMaintenance(t *testing.T) {
    db := testDB(t)

    m := &models.Maintenance{Description: "original", ReferenceType: "Space", SpaceType: "Yard", Date: "2026-01-01", Cost: 10.0, Notes: "old notes"}
    added, err := AddMaintenance(db, m)
    if err != nil {
        t.Fatalf("AddMaintenance failed: %v", err)
    }

    updated, err := UpdateMaintenance(db, added.ID, "updated desc", "2026-06-01", 99.99, "new notes")
    if err != nil {
        t.Fatalf("UpdateMaintenance failed: %v", err)
    }
    if updated.Description != "updated desc" {
        t.Errorf("expected description 'updated desc', got %q", updated.Description)
    }
    if updated.Date != "2026-06-01" {
        t.Errorf("expected date '2026-06-01', got %q", updated.Date)
    }
    if updated.Cost != 99.99 {
        t.Errorf("expected cost 99.99, got %v", updated.Cost)
    }
    if updated.Notes != "new notes" {
        t.Errorf("expected notes 'new notes', got %q", updated.Notes)
    }

    // Verify persisted
    got, err := GetMaintenance(db, added.ID)
    if err != nil {
        t.Fatalf("GetMaintenance failed: %v", err)
    }
    if got.Description != "updated desc" {
        t.Errorf("persisted description mismatch: %q", got.Description)
    }
}
