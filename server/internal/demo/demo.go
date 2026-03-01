package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// DemoData is the shape of sample_data.json
type DemoData struct {
    Appliances   []models.Appliance `json:"appliances"`
    Todos        []struct {
        Label         string  `json:"label"`
        Checked       bool    `json:"checked"`
        UserID        string  `json:"userID"`
        ApplianceIndex *int    `json:"applianceIndex,omitempty"`
        SpaceType     string  `json:"spaceType,omitempty"`
    } `json:"todos"`
    Notes        []struct {
        Title         string `json:"title"`
        Body          string `json:"body"`
        ApplianceIndex *int   `json:"applianceIndex,omitempty"`
        SpaceType     string `json:"spaceType,omitempty"`
    } `json:"notes"`
    Maintenances []struct {
        Description   string  `json:"description"`
        Date          string  `json:"date"`
        Cost          float64 `json:"cost"`
        Notes         string  `json:"notes"`
        ReferenceType string  `json:"referenceType"`
        SpaceType     string  `json:"spaceType"`
        ApplianceIndex *int    `json:"applianceIndex,omitempty"`
    } `json:"maintenances"`
    Repairs []struct {
        Description   string  `json:"description"`
        Date          string  `json:"date"`
        Cost          float64 `json:"cost"`
        Notes         string  `json:"notes"`
        ReferenceType string  `json:"referenceType"`
        SpaceType     string  `json:"spaceType"`
        ApplianceIndex *int    `json:"applianceIndex,omitempty"`
    } `json:"repairs"`
    Files []struct {
        OriginalName  string `json:"originalName"`
        UserID        string `json:"userID"`
        ApplianceIndex *int   `json:"applianceIndex,omitempty"`
        MaintenanceIndex *int `json:"maintenanceIndex,omitempty"`
        RepairIndex    *int   `json:"repairIndex,omitempty"`
        SpaceType      string `json:"spaceType,omitempty"`
    } `json:"files"`
}

// Seed loads the demo JSON from the provided file path (or default) and inserts data into the DB.
// Non-fatal errors are logged.
func Seed(db *gorm.DB, demoFilePath string) error {
    filePath := demoFilePath
    if filePath == "" {
        filePath = filepath.Join("server", "internal", "demo", "sample_data.json")
    }
    b, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("reading demo data from %s: %w", filePath, err)
    }

    var d DemoData
    if err := json.Unmarshal(b, &d); err != nil {
        return fmt.Errorf("unmarshal demo data: %w", err)
    }

    // create appliances and keep ids
    applianceIDs := make([]uint, len(d.Appliances))
    for i, a := range d.Appliances {
        // ensure zero ID so DB will assign
        a.ID = 0
        created, err := database.AddAppliance(db, &a)
        if err != nil {
            fmt.Printf("demo: error creating appliance %d: %v\n", i, err)
            continue
        }
        applianceIDs[i] = created.ID
    }

    // todos
    for i, t := range d.Todos {
        var aid uint = 0
        if t.ApplianceIndex != nil {
            idx := *t.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                aid = applianceIDs[idx]
            }
        }
        if _, err := database.AddTodo(db, t.Label, t.Checked, t.UserID, aid, t.SpaceType); err != nil {
            fmt.Printf("demo: error adding todo %d: %v\n", i, err)
        }
    }

    // notes
    for i, n := range d.Notes {
        var aid uint = 0
        if n.ApplianceIndex != nil {
            idx := *n.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                aid = applianceIDs[idx]
            }
        }
        if _, err := database.AddNote(db, n.Title, n.Body, aid, n.SpaceType); err != nil {
            fmt.Printf("demo: error adding note %d: %v\n", i, err)
        }
    }

    // maintenances
    maintenanceIDs := make([]uint, 0, len(d.Maintenances))
    for i, m := range d.Maintenances {
        var aid *uint
        if m.ApplianceIndex != nil {
            idx := *m.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                v := applianceIDs[idx]
                aid = &v
            }
        }
        mm := &models.Maintenance{
            Description:   m.Description,
            Date:          m.Date,
            Cost:          m.Cost,
            Notes:         m.Notes,
            SpaceType:     m.SpaceType,
            ReferenceType: m.ReferenceType,
            ApplianceID:   aid,
        }
        created, err := database.AddMaintenance(db, mm)
        if err != nil {
            fmt.Printf("demo: error adding maintenance %d: %v\n", i, err)
            continue
        }
        maintenanceIDs = append(maintenanceIDs, created.ID)
    }

    // repairs
    repairIDs := make([]uint, 0, len(d.Repairs))
    for i, r := range d.Repairs {
        var aid *uint
        if r.ApplianceIndex != nil {
            idx := *r.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                v := applianceIDs[idx]
                aid = &v
            }
        }
        rr := &models.Repair{
            Description:   r.Description,
            Date:          r.Date,
            Cost:          r.Cost,
            Notes:         r.Notes,
            SpaceType:     r.SpaceType,
            ReferenceType: r.ReferenceType,
            ApplianceID:   aid,
        }
        created, err := database.AddRepair(db, rr)
        if err != nil {
            fmt.Printf("demo: error adding repair %d: %v\n", i, err)
            continue
        }
        repairIDs = append(repairIDs, created.ID)
    }

    // files
    for i, f := range d.Files {
        sf := &models.SavedFile{OriginalName: f.OriginalName, UserID: f.UserID}
        created, err := database.UploadFile(db, sf)
        if err != nil {
            fmt.Printf("demo: error uploading file %d: %v\n", i, err)
            continue
        }

        // set a reasonable path so other code can reference it
        uploadsBase := filepath.Join("data", "uploads")
        if os.Getenv("DEMO_MODE") == "true" || os.Getenv("DEMO_MODE") == "1" {
            uploadsBase = filepath.Join(uploadsBase, "demo-uploads")
        }
        path := filepath.Join(uploadsBase, fmt.Sprintf("%d", created.ID))
        created.Path = path
        if _, err := database.UpdateFilePath(db, created); err != nil {
            fmt.Printf("demo: error updating file path %d: %v\n", i, err)
        }

        if f.ApplianceIndex != nil {
            idx := *f.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                _ = database.AttachFileToAppliance(db, created.ID, applianceIDs[idx])
            }
        }
        if f.MaintenanceIndex != nil {
            idx := *f.MaintenanceIndex
            if idx >= 0 && idx < len(maintenanceIDs) {
                _ = database.AttachFileToMaintenance(db, created.ID, maintenanceIDs[idx])
            }
        }
        if f.RepairIndex != nil {
            idx := *f.RepairIndex
            if idx >= 0 && idx < len(repairIDs) {
                _ = database.AttachFileToRepair(db, created.ID, repairIDs[idx])
            }
        }
        if f.SpaceType != "" {
            _ = database.AttachFileToSpace(db, created.ID, f.SpaceType)
        }
    }

    fmt.Println("demo: seeding complete")
    return nil
}
