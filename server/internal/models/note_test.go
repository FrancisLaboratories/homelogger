package models

import (
    "testing"
)

func TestNoteGormCRUD(t *testing.T) {
    db := openInMemory(t)

    if err := db.AutoMigrate(&Note{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    st := "Kitchen"
    n := &Note{Title: "T1", Body: "body", SpaceType: &st}
    if err := db.Create(n).Error; err != nil {
        t.Fatalf("create note: %v", err)
    }
    if n.ID == 0 {
        t.Fatalf("expected note ID after create")
    }

    var got Note
    if err := db.First(&got, n.ID).Error; err != nil {
        t.Fatalf("first note: %v", err)
    }
    if got.Title != n.Title {
        t.Fatalf("title mismatch: %s != %s", got.Title, n.Title)
    }

    // query by space type
    var notes []Note
    if err := db.Where("space_type = ?", st).Find(&notes).Error; err != nil {
        t.Fatalf("query notes: %v", err)
    }
    if len(notes) == 0 {
        t.Fatalf("expected at least one note for space %s", st)
    }

    // delete
    if err := db.Delete(&Note{}, n.ID).Error; err != nil {
        t.Fatalf("delete note: %v", err)
    }
}
