package database

import (
    "os"
    "testing"

    "github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestUploadGetUpdateDeleteFile(t *testing.T) {
    db := testDB(t)

    // create a temp file to simulate disk attachment
    tmp, err := os.CreateTemp(os.TempDir(), "hl-test-*")
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }
    tmpPath := tmp.Name()
    tmp.Close()

    // Upload file record pointing to temp path
    f := &models.SavedFile{Path: tmpPath, OriginalName: "orig.txt", UserID: "u1"}
    up, err := UploadFile(db, f)
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // GetFileInfo
    info, err := GetFileInfo(db, up.ID)
    if err != nil {
        t.Fatalf("GetFileInfo failed: %v", err)
    }
    if info.OriginalName != "orig.txt" || info.UserID != "u1" {
        t.Fatalf("unexpected file info: %#v", info)
    }

    // GetFilePath
    path, err := GetFilePath(db, up.ID)
    if err != nil {
        t.Fatalf("GetFilePath failed: %v", err)
    }
    if path != tmpPath {
        t.Fatalf("file path mismatch: %s vs %s", path, tmpPath)
    }

    // Update path
    up.Path = tmpPath + ".new"
    if _, err := UpdateFilePath(db, up); err != nil {
        t.Fatalf("UpdateFilePath failed: %v", err)
    }

    // Delete file record (hard delete)
    if err := DeleteFile(db, up.ID); err != nil {
        t.Fatalf("DeleteFile failed: %v", err)
    }
}

func TestAttachAndGetFilesByReferences(t *testing.T) {
    db := testDB(t)

    // create maintenance and repair
    m := &models.Maintenance{Description: "m1", ReferenceType: "Appliance", SpaceType: "", Date: "2026-01-01"}
    if _, err := AddMaintenance(db, m); err != nil {
        t.Fatalf("AddMaintenance failed: %v", err)
    }
    r := &models.Repair{Description: "r1", ReferenceType: "Appliance", SpaceType: "", Date: "2026-01-02"}
    if _, err := AddRepair(db, r); err != nil {
        t.Fatalf("AddRepair failed: %v", err)
    }

    // create temp file
    tmp, err := os.CreateTemp(os.TempDir(), "hl-test-*")
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }
    path := tmp.Name()
    tmp.Close()

    f := &models.SavedFile{Path: path, OriginalName: "o.txt", UserID: "u"}
    up, err := UploadFile(db, f)
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // Attach to maintenance
    if err := AttachFileToMaintenance(db, up.ID, m.ID); err != nil {
        t.Fatalf("AttachFileToMaintenance failed: %v", err)
    }
    files, err := GetFilesByMaintenance(db, m.ID)
    if err != nil {
        t.Fatalf("GetFilesByMaintenance failed: %v", err)
    }
    if len(files) != 1 || files[0].ID != up.ID {
        t.Fatalf("unexpected files by maintenance: %#v", files)
    }

    // Attach to repair
    // first upload another file
    tmp2, _ := os.CreateTemp(os.TempDir(), "hl-test-*")
    tmp2Path := tmp2.Name()
    tmp2.Close()
    f2 := &models.SavedFile{Path: tmp2Path, OriginalName: "o2.txt", UserID: "u"}
    up2, err := UploadFile(db, f2)
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }
    if err := AttachFileToRepair(db, up2.ID, r.ID); err != nil {
        t.Fatalf("AttachFileToRepair failed: %v", err)
    }
    rf, err := GetFilesByRepair(db, r.ID)
    if err != nil {
        t.Fatalf("GetFilesByRepair failed: %v", err)
    }
    if len(rf) != 1 || rf[0].ID != up2.ID {
        t.Fatalf("unexpected files by repair: %#v", rf)
    }
}
