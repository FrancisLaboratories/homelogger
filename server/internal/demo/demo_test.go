package demo

import (
	"testing"
	"time"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openInMemoryDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file:mem_demo_demo?mode=memory&cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open in-memory db: %v", err)
    }
    if err := database.MigrateGorm(db); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    return db
}

func TestSeedLoadsSampleData(t *testing.T) {
    db := openInMemoryDB(t)

    // sample_data.json lives in the same package directory as this test,
    // so reference it relative to the package (working dir during `go test`).
    demoPath := "sample_data.json"
    if err := Seed(db, demoPath); err != nil {
        t.Fatalf("Seed returned error: %v", err)
    }

    // appliances
    apps, err := database.GetAppliances(db)
    if err != nil {
        t.Fatalf("GetAppliances error: %v", err)
    }
    if len(apps) != 2 {
        t.Fatalf("expected 2 appliances from demo data, got %d", len(apps))
    }

    // tasks for first appliance should include a known label
    aid := apps[0].ID
    tasks, err := database.GetTasks(db, aid, "", true)
    if err != nil {
        t.Fatalf("GetTasks error: %v", err)
    }
    if len(tasks) == 0 {
        t.Fatalf("expected tasks for appliance %d, got 0", aid)
    }

    for _, tk := range tasks {
        if tk.DueDate == nil {
            continue
        }
        parsed, err := time.Parse("2006-01-02", *tk.DueDate)
        if err != nil {
            t.Fatalf("task %q has unparseable dueDate %q: %v", tk.Label, *tk.DueDate, err)
        }
        if tk.Label == "Replace central HVAC filter" {
            if parsed.Year() != 2026 || parsed.Month() != 6 || parsed.Day() != 1 {
                t.Fatalf("overdue task dueDate expected 2026-06-01, got %s", *tk.DueDate)
            }
            if !parsed.Before(time.Now()) {
                t.Fatal("overdue task should be in the past")
            }
        } else if parsed.Year() < 2026 || parsed.Year() > 2027 {
            t.Fatalf("task %q dueDate %s out of expected range (2026-2027)", tk.Label, *tk.DueDate)
        }
    }

    // notes for first appliance
    notes, err := database.GetNotes(db, aid, "")
    if err != nil {
        t.Fatalf("GetNotes error: %v", err)
    }
    if len(notes) == 0 {
        t.Fatalf("expected notes for appliance %d, got 0", aid)
    }

    // maintenances and repairs should exist for appliance 0
    maint, err := database.GetMaintenances(db, aid, "Appliance", "")
    if err != nil {
        t.Fatalf("GetMaintenances error: %v", err)
    }
    if len(maint) == 0 {
        t.Fatalf("expected maintenances for appliance %d, got 0", aid)
    }

    rep, err := database.GetRepairs(db, aid, "Appliance", "")
    if err != nil {
        t.Fatalf("GetRepairs error: %v", err)
    }
    if len(rep) == 0 {
        t.Fatalf("expected repairs for appliance %d, got 0", aid)
    }
}

func TestShiftDates(t *testing.T) {
    may01 := "2026-05-01"
    jun15 := "2026-06-15"
    apr15 := "2026-04-15"
    apr10 := "2026-04-10"
    jul01 := "2026-07-01"
    sep01 := "2026-09-01"
    maintOld := "2023-11-20"
    repairOld := "2020-07-22"

    d := DemoData{
        Tasks: []struct {
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
        }{
            {Label: "Task 0", DueDate: &may01},
            {Label: "Task 1", DueDate: &jun15},
            {Label: "Task 2", DueDate: &apr15},
            {Label: "Task 3", DueDate: &apr15},
            {Label: "Replace central HVAC filter", DueDate: &apr10, Priority: "high"},
            {Label: "Task 5", DueDate: &jul01},
            {Label: "Task 6", DueDate: &may01},
            {Label: "Nil due", DueDate: nil},
            {Label: "Task 8", DueDate: &jun15},
            {Label: "Task 10", DueDate: &sep01},
        },
        Maintenances: []struct {
            Description   string  `json:"description"`
            Date          string  `json:"date"`
            Cost          float64 `json:"cost"`
            Notes         string  `json:"notes"`
            ReferenceType string  `json:"referenceType"`
            SpaceType     string  `json:"spaceType"`
            ApplianceIndex *int    `json:"applianceIndex,omitempty"`
        }{
            {Description: "Washer belt replacement", Date: maintOld, Cost: 120},
        },
        Repairs: []struct {
            Description   string  `json:"description"`
            Date          string  `json:"date"`
            Cost          float64 `json:"cost"`
            Notes         string  `json:"notes"`
            ReferenceType string  `json:"referenceType"`
            SpaceType     string  `json:"spaceType"`
            ApplianceIndex *int    `json:"applianceIndex,omitempty"`
        }{
            {Description: "Washer motor repair", Date: repairOld, Cost: 250},
        },
    }

    d.shiftDates()

    // overdue task (index 4) hardcoded to 2026-06-01
    if d.Tasks[4].DueDate == nil {
        t.Fatal("overdue task dueDate should not be nil")
    }
    if *d.Tasks[4].DueDate != "2026-06-01" {
        t.Fatalf("overdue task expected 2026-06-01, got %s", *d.Tasks[4].DueDate)
    }

    // nil dueDate tasks stay nil
    if d.Tasks[7].DueDate != nil {
        t.Fatalf("task with nil dueDate should stay nil, got %s", *d.Tasks[7].DueDate)
    }

    // non-overdue tasks with dueDate should be parseable and in 2026–2027
    now := time.Now()
    for i, tk := range d.Tasks {
        if tk.DueDate == nil || i == 4 {
            continue
        }
        parsed, err := time.Parse("2006-01-02", *tk.DueDate)
        if err != nil {
            t.Fatalf("task %d (%q) unparseable date %q: %v", i, tk.Label, *tk.DueDate, err)
        }
        if parsed.Year() < 2026 || parsed.Year() > 2027 {
            t.Fatalf("task %d (%q) year %d out of 2026–2027", i, tk.Label, parsed.Year())
        }
        if !parsed.After(now) {
            t.Fatalf("non-overdue task %d (%q) date %s should be in the future", i, tk.Label, *tk.DueDate)
        }
    }

    // maintenance shifted +1 year
    maintParsed, err := time.Parse("2006-01-02", d.Maintenances[0].Date)
    if err != nil {
        t.Fatalf("maintenance date unparseable: %v", err)
    }
    maintOrig, _ := time.Parse("2006-01-02", maintOld)
    expectedMaint := maintOrig.AddDate(1, 0, 0)
    if !maintParsed.Equal(expectedMaint) {
        t.Fatalf("maintenance expected %s, got %s", expectedMaint.Format("2006-01-02"), d.Maintenances[0].Date)
    }

    // repair shifted +1 year
    repairParsed, err := time.Parse("2006-01-02", d.Repairs[0].Date)
    if err != nil {
        t.Fatalf("repair date unparseable: %v", err)
    }
    repairOrig, _ := time.Parse("2006-01-02", repairOld)
    expectedRepair := repairOrig.AddDate(1, 0, 0)
    if !repairParsed.Equal(expectedRepair) {
        t.Fatalf("repair expected %s, got %s", expectedRepair.Format("2006-01-02"), d.Repairs[0].Date)
    }

    // verify daysInMonth clamping: task with day=31 shifted to month with ≤30 should clamp
    for _, tk := range d.Tasks {
        if tk.DueDate == nil {
            continue
        }
        parsed, _ := time.Parse("2006-01-02", *tk.DueDate)
        maxDay := time.Date(parsed.Year(), parsed.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
        if parsed.Day() > maxDay {
            t.Fatalf("task %q day %d exceeds max %d for month %s", tk.Label, parsed.Day(), maxDay, parsed.Month())
        }
    }
}

func TestSeedMissingFile(t *testing.T) {
    db := openInMemoryDB(t)
    if err := Seed(db, "nonexistent_demo_file.json"); err == nil {
        t.Fatalf("expected error when demo file missing")
    }
}
