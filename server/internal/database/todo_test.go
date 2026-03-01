package database

import (
	"testing"

	"github.com/masoncfrancis/homelogger/server/internal/models"
)

func TestAddAndGetTodos(t *testing.T) {
    db := testDB(t)

    // Add two todos: one general and one for an appliance
    todo1, err := AddTodo(db, "task1", false, "1", 0, "")
    if err != nil {
        t.Fatalf("AddTodo failed: %v", err)
    }

    // Create appliance and add a todo for it
    a := &models.Appliance{ApplianceName: "A", Manufacturer: "M", ModelNumber: "X", SerialNumber: "S", YearPurchased: "2020", PurchasePrice: "1", Location: "L", Type: "T"}
    if _, err := AddAppliance(db, a); err != nil {
        t.Fatalf("AddAppliance failed: %v", err)
    }
    _, err = AddTodo(db, "task2", true, "1", a.ID, "")
    if err != nil {
        t.Fatalf("AddTodo for appliance failed: %v", err)
    }

    // GetTodos with no filters (userID is fixed to "1" in code, but AddTodo uses provided userID)
    todos, err := GetTodos(db, 0, "")
    if err != nil {
        t.Fatalf("GetTodos failed: %v", err)
    }
    if len(todos) == 0 {
        t.Fatalf("expected some todos, got none")
    }

    // Change checked state and verify
    if err := ChangeTodoChecked(db, todo1.ID, true); err != nil {
        t.Fatalf("ChangeTodoChecked failed: %v", err)
    }
}

func TestDeleteTodo(t *testing.T) {
    db := testDB(t)

    td, err := AddTodo(db, "to-delete", false, "1", 0, "")
    if err != nil {
        t.Fatalf("AddTodo failed: %v", err)
    }

    if err := DeleteTodo(db, td.ID); err != nil {
        t.Fatalf("DeleteTodo failed: %v", err)
    }
}
func TestAddGetDeleteTodo(t *testing.T) {
    db := testDB(t)

    // add a todo (use userID "1" to match GetTodos filter)
    todo, err := AddTodo(db, "test task", false, "1", 0, "")
    if err != nil {
        t.Fatalf("AddTodo failed: %v", err)
    }
    if todo.Label != "test task" {
        t.Fatalf("unexpected label: %s", todo.Label)
    }

    // get todos
    todos, err := GetTodos(db, 0, "")
    if err != nil {
        t.Fatalf("GetTodos failed: %v", err)
    }
    if len(todos) != 1 {
        t.Fatalf("expected 1 todo, got %d", len(todos))
    }
    if todos[0].Label != "test task" {
        t.Fatalf("unexpected todo label: %s", todos[0].Label)
    }

    // delete todo
    if err := DeleteTodo(db, todo.ID); err != nil {
        t.Fatalf("DeleteTodo failed: %v", err)
    }

    // verify deletion
    todos, err = GetTodos(db, 0, "")
    if err != nil {
        t.Fatalf("GetTodos failed after delete: %v", err)
    }
    if len(todos) != 0 {
        t.Fatalf("expected 0 todos after delete, got %d", len(todos))
    }
}
