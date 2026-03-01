package database

import (
    "testing"

    "github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestAddGetUpdateDeleteNote(t *testing.T) {
    db := testDB(t)

    // Add
    n, err := AddNote(db, "title1", "body1", 0, "")
    if err != nil {
        t.Fatalf("AddNote failed: %v", err)
    }

    // Get
    got, err := GetNote(db, n.ID)
    if err != nil {
        t.Fatalf("GetNote failed: %v", err)
    }
    if got.Title != "title1" || got.Body != "body1" {
        t.Fatalf("unexpected note contents: %#v", got)
    }

    // Update
    upd, err := UpdateNote(db, n.ID, "newtitle", "newbody")
    if err != nil {
        t.Fatalf("UpdateNote failed: %v", err)
    }
    if upd.Title != "newtitle" || upd.Body != "newbody" {
        t.Fatalf("update didn't apply: %#v", upd)
    }

    // GetNotes (no filters)
    notes, err := GetNotes(db, 0, "")
    if err != nil {
        t.Fatalf("GetNotes failed: %v", err)
    }
    if len(notes) != 1 {
        t.Fatalf("expected 1 note, got %d", len(notes))
    }

    // Delete
    if err := DeleteNote(db, n.ID); err != nil {
        t.Fatalf("DeleteNote failed: %v", err)
    }

    // Ensure gone
    _, err = GetNote(db, n.ID)
    if err == nil {
        t.Fatalf("expected error when getting deleted note")
    }

    // Add with appliance and space filters
    a := &models.Appliance{ApplianceName: "A", Manufacturer: "M", ModelNumber: "X", SerialNumber: "S", YearPurchased: "2020", PurchasePrice: "1", Location: "L", Type: "T"}
    if _, err := AddAppliance(db, a); err != nil {
        t.Fatalf("AddAppliance failed: %v", err)
    }

    n2, err := AddNote(db, "t2", "b2", a.ID, "Kitchen")
    if err != nil {
        t.Fatalf("AddNote with appliance failed: %v", err)
    }
    // Filtered list
    ns, err := GetNotes(db, a.ID, "Kitchen")
    if err != nil {
        t.Fatalf("GetNotes filtered failed: %v", err)
    }
    if len(ns) != 1 || ns[0].ID != n2.ID {
        t.Fatalf("filtered notes mismatch: %#v", ns)
    }
}
