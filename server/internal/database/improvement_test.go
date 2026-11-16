package database

import (
	"os"
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestImprovementCRUDAndSpaceFilter(t *testing.T) {
	db := TestDB(t)
	appliance, err := AddAppliance(db, &models.Appliance{ApplianceName: "HVAC"})
	if err != nil {
		t.Fatalf("AddAppliance: %v", err)
	}
	improvement := &models.Improvement{
		Description:   "Install smart thermostat",
		Date:          "2026-01-01",
		Cost:          250,
		Notes:         "Initial notes",
		ReferenceType: "Space",
		SpaceType:     "HVAC",
	}
	created, err := AddImprovement(db, improvement)
	if err != nil {
		t.Fatalf("AddImprovement: %v", err)
	}

	records, err := GetImprovements(db, 0, "Space", "HVAC")
	if err != nil {
		t.Fatalf("GetImprovements: %v", err)
	}
	if len(records) != 1 || records[0].ID != created.ID {
		t.Fatalf("expected created improvement in HVAC results, got %#v", records)
	}

	applianceImprovement, err := AddImprovement(db, &models.Improvement{
		Description:   "Replace control panel",
		Date:          "2026-01-02",
		ReferenceType: "Appliance",
		ApplianceID:   &appliance.ID,
	})
	if err != nil {
		t.Fatalf("AddImprovement for appliance: %v", err)
	}
	records, err = GetImprovements(db, appliance.ID, "Appliance", "")
	if err != nil || len(records) != 1 || records[0].ID != applianceImprovement.ID {
		t.Fatalf("expected appliance improvement, got %#v (err %v)", records, err)
	}

	updated, err := UpdateImprovement(db, created.ID, "Updated thermostat", "2026-02-01", 300, "Updated notes")
	if err != nil {
		t.Fatalf("UpdateImprovement: %v", err)
	}
	if updated.Description != "Updated thermostat" || updated.Cost != 300 || updated.Notes != "Updated notes" {
		t.Fatalf("unexpected updated record: %#v", updated)
	}
}

func TestDeleteImprovementRemovesAttachments(t *testing.T) {
	db := TestDB(t)
	improvement, err := AddImprovement(db, &models.Improvement{
		Description:   "Upgrade lighting",
		Date:          "2026-01-01",
		ReferenceType: "Space",
		SpaceType:     "Yard",
	})
	if err != nil {
		t.Fatalf("AddImprovement: %v", err)
	}

	file, err := os.CreateTemp(os.TempDir(), "hl-improvement-*")
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		t.Fatalf("close attachment: %v", err)
	}
	savedFile, err := UploadFile(db, &models.SavedFile{
		Path: path, OriginalName: "receipt.pdf", UserID: "test",
	})
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if err := AttachFileToImprovement(db, savedFile.ID, improvement.ID); err != nil {
		t.Fatalf("AttachFileToImprovement: %v", err)
	}

	files, err := GetFilesByImprovement(db, improvement.ID)
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one attached file, got %#v (err %v)", files, err)
	}
	if err := DeleteImprovement(db, improvement.ID); err != nil {
		t.Fatalf("DeleteImprovement: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected attachment file removed, stat error: %v", err)
	}
	files, err = GetFilesByImprovement(db, improvement.ID)
	if err != nil || len(files) != 0 {
		t.Fatalf("expected attachment metadata removed, got %#v (err %v)", files, err)
	}
}
