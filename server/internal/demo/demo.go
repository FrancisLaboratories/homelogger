package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// DemoData is the shape of sample_data.json
type DemoData struct {
    Appliances []models.Appliance `json:"appliances"`
    Tasks      []struct {
        Label              string   `json:"label"`
        Notes              string   `json:"notes,omitempty"`
        Checked            bool     `json:"checked"`
        Priority           string   `json:"priority,omitempty"`
        DueDate            *string  `json:"dueDate,omitempty"`
        EstimatedCost      *float64 `json:"estimatedCost,omitempty"`
        IsRecurring        bool     `json:"isRecurring"`
        RecurrenceInterval int      `json:"recurrenceInterval,omitempty"`
        RecurrenceUnit     string   `json:"recurrenceUnit,omitempty"`
        RecurrenceMode     string   `json:"recurrenceMode,omitempty"`
        UserID             string   `json:"userID"`
        ApplianceIndex     *int     `json:"applianceIndex,omitempty"`
        SpaceType          string   `json:"spaceType,omitempty"`
    } `json:"tasks"`
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

func daysInMonth(m, y int) int {
    return time.Date(y, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// shiftDates transforms all dates in demo data so they appear current.
// Task dueDates spread across Aug 2026 – Jan 2027, with one overdue.
// Maintenance and repair dates shift +1 year so all stay historical.
func (d *DemoData) shiftDates() {
    now := time.Now()

    for i := range d.Tasks {
        if d.Tasks[i].DueDate == nil {
            continue
        }
        if i == 4 {
            s := "2026-06-01"
            d.Tasks[i].DueDate = &s
            continue
        }
        orig, err := time.Parse("2006-01-02", *d.Tasks[i].DueDate)
        if err != nil {
            continue
        }
        monthOffset := 1 + (i % 6)
        targetMonth := int(now.Month()) + monthOffset
        targetYear := now.Year()
        if targetMonth > 12 {
            targetMonth -= 12
            targetYear++
        }
        day := orig.Day()
        if max := daysInMonth(targetMonth, targetYear); day > max {
            day = max
        }
        s := fmt.Sprintf("%d-%02d-%02d", targetYear, targetMonth, day)
        d.Tasks[i].DueDate = &s
    }

    for i := range d.Maintenances {
        orig, err := time.Parse("2006-01-02", d.Maintenances[i].Date)
        if err != nil {
            continue
        }
        d.Maintenances[i].Date = orig.AddDate(1, 0, 0).Format("2006-01-02")
    }

    for i := range d.Repairs {
        orig, err := time.Parse("2006-01-02", d.Repairs[i].Date)
        if err != nil {
            continue
        }
        d.Repairs[i].Date = orig.AddDate(1, 0, 0).Format("2006-01-02")
    }
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
    d.shiftDates()

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

    // tasks
    for i, t := range d.Tasks {
        task := &models.Task{
            Label:              t.Label,
            Notes:              t.Notes,
            Checked:            t.Checked,
            Priority:           t.Priority,
            DueDate:            t.DueDate,
            EstimatedCost:      t.EstimatedCost,
            IsRecurring:        t.IsRecurring,
            RecurrenceInterval: t.RecurrenceInterval,
            RecurrenceUnit:     t.RecurrenceUnit,
            RecurrenceMode:     t.RecurrenceMode,
            UserID:             t.UserID,
        }
        if t.ApplianceIndex != nil {
            idx := *t.ApplianceIndex
            if idx >= 0 && idx < len(applianceIDs) {
                aid := applianceIDs[idx]
                task.ApplianceID = &aid
            }
        }
        if t.SpaceType != "" {
            st := t.SpaceType
            task.SpaceType = &st
        }
        if _, err := database.AddTask(db, task); err != nil {
            fmt.Printf("demo: error adding task %d: %v\n", i, err)
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
