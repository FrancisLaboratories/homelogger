package demo

import (
	"testing"
	"time"

	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	dateMay01 = "2026-05-01"
	dateJun15 = "2026-06-15"
	dateApr15 = "2026-04-15"
	dateMaint = "2023-11-20"
	dateRep   = "2020-07-22"
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

    verifySeedTaskDates(t, tasks)

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

type testTask struct {
    Label              string
    DueDate            *string
    Priority           string
}
type testMaintenance struct {
    Description string
    Date        string
    Cost        float64
}
type testRepair struct {
    Description string
    Date        string
    Cost        float64
}

func demoDataFromFixtures(tasks []testTask, maintenances []testMaintenance, repairs []testRepair) DemoData {
    d := DemoData{}
    for _, tt := range tasks {
        st := tt
        d.Tasks = append(d.Tasks, struct {
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
        }{Label: st.Label, DueDate: st.DueDate, Priority: st.Priority})
    }
    for _, mt := range maintenances {
        sm := mt
        d.Maintenances = append(d.Maintenances, struct {
            Description   string  `json:"description"`
            Date          string  `json:"date"`
            Cost          float64 `json:"cost"`
            Notes         string  `json:"notes"`
            ReferenceType string  `json:"referenceType"`
            SpaceType     string  `json:"spaceType"`
            ApplianceIndex *int    `json:"applianceIndex,omitempty"`
        }{Description: sm.Description, Date: sm.Date, Cost: sm.Cost})
    }
    for _, rt := range repairs {
        sr := rt
        d.Repairs = append(d.Repairs, struct {
            Description   string  `json:"description"`
            Date          string  `json:"date"`
            Cost          float64 `json:"cost"`
            Notes         string  `json:"notes"`
            ReferenceType string  `json:"referenceType"`
            SpaceType     string  `json:"spaceType"`
            ApplianceIndex *int    `json:"applianceIndex,omitempty"`
        }{Description: sr.Description, Date: sr.Date, Cost: sr.Cost})
    }
    return d
}

func verifyDateShift(t *testing.T, label, got, want string) {
    t.Helper()
    if got != want {
        t.Fatalf("%s expected %s, got %s", label, want, got)
    }
}

func verifyTaskDueDateAfter(t *testing.T, label, dueDate string, ref time.Time) {
    t.Helper()
    parsed, err := time.Parse("2006-01-02", dueDate)
    if err != nil {
        t.Fatalf("task %q unparseable date %q: %v", label, dueDate, err)
    }
    if parsed.Year() < 2026 || parsed.Year() > 2027 {
        t.Fatalf("task %q year %d out of 2026–2027", label, parsed.Year())
    }
    if !parsed.After(ref) {
        t.Fatalf("task %q date %s should be in the future", label, dueDate)
    }
}

func verifyDayClamping(t *testing.T, tasks []struct {
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
}) {
    for _, tk := range tasks {
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

func verifyOverdueTaskDueDate(t *testing.T, d *DemoData) {
    t.Helper()
    if d.Tasks[4].DueDate == nil {
        t.Fatal("overdue task dueDate should not be nil")
    }
    verifyDateShift(t, "overdue task", *d.Tasks[4].DueDate, overdueDate)
}

func verifySeedTaskDates(t *testing.T, tasks []models.Task) {
    t.Helper()
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
                t.Fatalf("overdue task dueDate expected %s, got %s", overdueDate, *tk.DueDate)
            }
            if !parsed.Before(time.Now()) {
                t.Fatal("overdue task should be in the past")
            }
        } else if parsed.Year() < 2026 || parsed.Year() > 2027 {
            t.Fatalf("task %q dueDate %s out of expected range (2026-2027)", tk.Label, *tk.DueDate)
        }
    }
}

func TestShiftDates(t *testing.T) {
    apr10 := "2026-04-10"
    jul01 := "2026-07-01"
    sep01 := "2026-09-01"

    d := demoDataFromFixtures(
        []testTask{
            {Label: "Task 0", DueDate: &dateMay01},
            {Label: "Task 1", DueDate: &dateJun15},
            {Label: "Task 2", DueDate: &dateApr15},
            {Label: "Task 3", DueDate: &dateApr15},
            {Label: "Replace central HVAC filter", DueDate: &apr10, Priority: "high"},
            {Label: "Task 5", DueDate: &jul01},
            {Label: "Task 6", DueDate: &dateMay01},
            {Label: "Nil due", DueDate: nil},
            {Label: "Task 8", DueDate: &dateJun15},
            {Label: "Task 10", DueDate: &sep01},
        },
        []testMaintenance{
            {Description: "Washer belt replacement", Date: dateMaint, Cost: 120},
        },
        []testRepair{
            {Description: "Washer motor repair", Date: dateRep, Cost: 250},
        },
    )

    d.shiftDates()

    t.Run("overdue task hardcoded to 2026-06-01", func(t *testing.T) {
        verifyOverdueTaskDueDate(t, &d)
    })

    t.Run("nil dueDate stays nil", func(t *testing.T) {
        if d.Tasks[7].DueDate != nil {
            t.Fatalf("task with nil dueDate should stay nil, got %s", *d.Tasks[7].DueDate)
        }
    })

    t.Run("non-overdue tasks in 2026–2027 and future", func(t *testing.T) {
        now := time.Now()
        for i, tk := range d.Tasks {
            if tk.DueDate == nil || i == 4 {
                continue
            }
            verifyTaskDueDateAfter(t, tk.Label, *tk.DueDate, now)
        }
    })

    t.Run("maintenance shifted +1 year", func(t *testing.T) {
        maintOrig, _ := time.Parse("2006-01-02", dateMaint)
        want := maintOrig.AddDate(1, 0, 0).Format("2006-01-02")
        verifyDateShift(t, "maintenance", d.Maintenances[0].Date, want)
    })

    t.Run("repair shifted +1 year", func(t *testing.T) {
        repairOrig, _ := time.Parse("2006-01-02", dateRep)
        want := repairOrig.AddDate(1, 0, 0).Format("2006-01-02")
        verifyDateShift(t, "repair", d.Repairs[0].Date, want)
    })

    t.Run("day clamping", func(t *testing.T) {
        verifyDayClamping(t, d.Tasks)
    })
}

func TestSeedMissingFile(t *testing.T) {
    db := openInMemoryDB(t)
    if err := Seed(db, "nonexistent_demo_file.json"); err == nil {
        t.Fatalf("expected error when demo file missing")
    }
}
