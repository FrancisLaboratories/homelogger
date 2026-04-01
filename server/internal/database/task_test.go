package database

import (
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func strPtr(s string) *string  { return &s }
func f64Ptr(f float64) *float64 { return &f }
func uintPtr(u uint) *uint     { return &u }

func TestAddAndGetTask(t *testing.T) {
	db := setupTestDB(t)

	task := &models.Task{
		Label:    "Replace HVAC filter",
		Notes:    "Check size first",
		Priority: "high",
		UserID:   "1",
	}

	created, err := AddTask(db, task)
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after add")
	}

	fetched, err := GetTask(db, created.ID)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if fetched.Label != task.Label {
		t.Fatalf("expected label %q, got %q", task.Label, fetched.Label)
	}
	if fetched.Priority != "high" {
		t.Fatalf("expected priority high, got %q", fetched.Priority)
	}
}

func TestGetTasksFilter(t *testing.T) {
	db := setupTestDB(t)

	spaceHVAC := "HVAC"
	spacePlumbing := "Plumbing"

	_, _ = AddTask(db, &models.Task{Label: "HVAC task", SpaceType: &spaceHVAC, UserID: "1"})
	_, _ = AddTask(db, &models.Task{Label: "Plumbing task", SpaceType: &spacePlumbing, UserID: "1"})

	hvacTasks, err := GetTasks(db, 0, "HVAC", false)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	if len(hvacTasks) != 1 {
		t.Fatalf("expected 1 HVAC task, got %d", len(hvacTasks))
	}
	if hvacTasks[0].Label != "HVAC task" {
		t.Fatalf("unexpected task label: %s", hvacTasks[0].Label)
	}
}

func TestGetAllActiveTasks(t *testing.T) {
	db := setupTestDB(t)

	spaceType := "Yard"
	_, _ = AddTask(db, &models.Task{Label: "Mow lawn", SpaceType: &spaceType, UserID: "1"})
	_, _ = AddTask(db, &models.Task{Label: "Check fence", SpaceType: &spaceType, UserID: "1"})

	tasks, err := GetAllActiveTasks(db)
	if err != nil {
		t.Fatalf("GetAllActiveTasks error: %v", err)
	}
	if len(tasks) < 2 {
		t.Fatalf("expected at least 2 active tasks, got %d", len(tasks))
	}
}

func TestCompleteNonRecurringTask(t *testing.T) {
	db := setupTestDB(t)

	due := "2026-04-01"
	created, err := AddTask(db, &models.Task{
		Label:   "One-time task",
		DueDate: &due,
		UserID:  "1",
	})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	completed, err := CompleteTask(db, created.ID, "2026-03-31")
	if err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	if !completed.Checked {
		t.Fatal("expected task to be marked checked after completion")
	}
	if completed.LastCompletedAt == nil || *completed.LastCompletedAt != "2026-03-31" {
		t.Fatalf("expected LastCompletedAt=2026-03-31, got %v", completed.LastCompletedAt)
	}

	// Should not appear in active tasks
	active, _ := GetAllActiveTasks(db)
	for _, t2 := range active {
		if t2.ID == created.ID {
			t.Fatal("completed task should not appear in active tasks")
		}
	}
}

func TestCompleteRecurringTask_CompletionDateMode(t *testing.T) {
	db := setupTestDB(t)

	due := "2026-04-01"
	created, err := AddTask(db, &models.Task{
		Label:              "Monthly filter check",
		DueDate:            &due,
		IsRecurring:        true,
		RecurrenceInterval: 1,
		RecurrenceUnit:     "months",
		RecurrenceMode:     "completion_date",
		UserID:             "1",
	})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	completed, err := CompleteTask(db, created.ID, "2026-03-15")
	if err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	// For completion_date mode, new due date = completion date + 1 month = 2026-04-15
	if completed.DueDate == nil || *completed.DueDate != "2026-04-15" {
		t.Fatalf("expected next due date 2026-04-15, got %v", completed.DueDate)
	}
	if completed.Checked {
		t.Fatal("recurring task should not be marked checked after completion")
	}
}

func TestCompleteRecurringTask_DueDateMode(t *testing.T) {
	db := setupTestDB(t)

	due := "2026-04-01"
	created, err := AddTask(db, &models.Task{
		Label:              "Quarterly inspection",
		DueDate:            &due,
		IsRecurring:        true,
		RecurrenceInterval: 3,
		RecurrenceUnit:     "months",
		RecurrenceMode:     "due_date",
		UserID:             "1",
	})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	completed, err := CompleteTask(db, created.ID, "2026-03-15")
	if err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	// For due_date mode, new due date = original due date + 3 months = 2026-07-01
	if completed.DueDate == nil || *completed.DueDate != "2026-07-01" {
		t.Fatalf("expected next due date 2026-07-01, got %v", completed.DueDate)
	}
}

func TestUpdateTask(t *testing.T) {
	db := setupTestDB(t)

	created, err := AddTask(db, &models.Task{Label: "Fix gutter", UserID: "1"})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	created.Label = "Clean gutter"
	created.Priority = "medium"
	updated, err := UpdateTask(db, created)
	if err != nil {
		t.Fatalf("UpdateTask error: %v", err)
	}
	if updated.Label != "Clean gutter" {
		t.Fatalf("expected updated label, got %q", updated.Label)
	}
	if updated.Priority != "medium" {
		t.Fatalf("expected updated priority, got %q", updated.Priority)
	}
}

func TestDeleteTask(t *testing.T) {
	db := setupTestDB(t)

	created, err := AddTask(db, &models.Task{Label: "Temp task", UserID: "1"})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	if err := DeleteTask(db, created.ID); err != nil {
		t.Fatalf("DeleteTask error: %v", err)
	}

	if _, err := GetTask(db, created.ID); err == nil {
		t.Fatal("expected error fetching deleted task")
	}
}

func TestUncompleteTask(t *testing.T) {
	db := setupTestDB(t)

	created, err := AddTask(db, &models.Task{Label: "Paint fence", UserID: "1"})
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}

	// Complete it first
	if _, err := CompleteTask(db, created.ID, "2026-03-31"); err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	// Now uncomplete
	undone, err := UncompleteTask(db, created.ID)
	if err != nil {
		t.Fatalf("UncompleteTask error: %v", err)
	}
	if undone.Checked {
		t.Fatal("expected task to be unchecked after uncomplete")
	}
}
