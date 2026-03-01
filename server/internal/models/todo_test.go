package models

import (
    "testing"
)

func TestTodoGormCRUD(t *testing.T) {
    db := openInMemory(t)

    if err := db.AutoMigrate(&Todo{}); err != nil {
        t.Fatalf("AutoMigrate: %v", err)
    }

    todo := &Todo{Label: "Do laundry", UserID: "1"}
    if err := db.Create(todo).Error; err != nil {
        t.Fatalf("create todo: %v", err)
    }
    if todo.ID == 0 {
        t.Fatalf("expected todo ID after create")
    }

    var got Todo
    if err := db.First(&got, todo.ID).Error; err != nil {
        t.Fatalf("first todo: %v", err)
    }
    if got.Checked {
        t.Fatalf("expected checked default false")
    }

    // toggle checked
    got.Checked = true
    if err := db.Save(&got).Error; err != nil {
        t.Fatalf("save todo: %v", err)
    }

    var after Todo
    if err := db.First(&after, todo.ID).Error; err != nil {
        t.Fatalf("first after update: %v", err)
    }
    if !after.Checked {
        t.Fatalf("expected checked true after update")
    }

    // query by user
    var todos []Todo
    if err := db.Where("user_id = ?", "1").Find(&todos).Error; err != nil {
        t.Fatalf("query todos: %v", err)
    }
    if len(todos) == 0 {
        t.Fatalf("expected todos for user 1")
    }

    // delete
    if err := db.Delete(&Todo{}, todo.ID).Error; err != nil {
        t.Fatalf("delete todo: %v", err)
    }
}
