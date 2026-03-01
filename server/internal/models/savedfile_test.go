package models

import (
    "testing"
)

func TestSavedFileGormCRUD(t *testing.T) {
    db := openInMemory(t)

    if err := db.AutoMigrate(&SavedFile{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    sf := &SavedFile{Path: "/tmp/file1", OriginalName: "file1.txt", UserID: "1"}
    if err := db.Create(sf).Error; err != nil {
        t.Fatalf("create savedfile: %v", err)
    }
    if sf.ID == 0 {
        t.Fatalf("expected savedfile ID after create")
    }

    var got SavedFile
    if err := db.First(&got, sf.ID).Error; err != nil {
        t.Fatalf("first savedfile: %v", err)
    }
    if got.Path != sf.Path {
        t.Fatalf("path mismatch: %s != %s", got.Path, sf.Path)
    }

    // update
    got.OriginalName = "renamed.txt"
    if err := db.Save(&got).Error; err != nil {
        t.Fatalf("save savedfile: %v", err)
    }

    var after SavedFile
    if err := db.First(&after, sf.ID).Error; err != nil {
        t.Fatalf("first after update: %v", err)
    }
    if after.OriginalName != "renamed.txt" {
        t.Fatalf("original name didn't update: %s", after.OriginalName)
    }

    // delete
    if err := db.Delete(&SavedFile{}, sf.ID).Error; err != nil {
        t.Fatalf("delete savedfile: %v", err)
    }
}
