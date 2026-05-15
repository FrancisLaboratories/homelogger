package database

import (
	"fmt"
	"time"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetTasks returns tasks filtered by optional applianceId and spaceType.
// Pass applianceId=0 and spaceType="" to get tasks with no filter.
// Set includeCompleted=true to include tasks where Checked=true.
func GetTasks(db *gorm.DB, applianceId uint, spaceType string, includeCompleted bool) ([]models.Task, error) {
	var tasks []models.Task
	query := db.Model(&models.Task{})

	if !includeCompleted {
		query = query.Where("checked = ?", false)
	}

	if applianceId != 0 {
		query = query.Where("appliance_id = ?", applianceId)
	} else if spaceType != "" {
		query = query.Where("space_type = ?", spaceType)
	} else {
		// Global (no appliance, no space)
		query = query.Where("appliance_id IS NULL AND space_type IS NULL")
	}

	result := query.Order("due_date ASC, created_at ASC").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// GetAllActiveTasks returns all incomplete tasks across all spaces and appliances,
// ordered by due date ascending (nulls last). Used for the dashboard.
func GetAllActiveTasks(db *gorm.DB) ([]models.Task, error) {
	return GetAllTasks(db, false)
}

// GetAllTasks returns tasks across all spaces and appliances.
// Pass includeCompleted=true to include tasks where checked=true.
func GetAllTasks(db *gorm.DB, includeCompleted bool) ([]models.Task, error) {
	var tasks []models.Task
	query := db.Model(&models.Task{})
	if !includeCompleted {
		query = query.Where("checked = ?", false)
	}
	result := query.
		Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, created_at ASC").
		Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// GetTask returns a single task by ID.
func GetTask(db *gorm.DB, id uint) (*models.Task, error) {
	var task models.Task
	result := db.Where("id = ?", id).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

// AddTask creates a new task record.
func AddTask(db *gorm.DB, task *models.Task) (*models.Task, error) {
	result := db.Create(task)
	if result.Error != nil {
		return nil, result.Error
	}
	return task, nil
}

// UpdateTask saves all fields of an existing task.
func UpdateTask(db *gorm.DB, task *models.Task) (*models.Task, error) {
	result := db.Save(task)
	if result.Error != nil {
		return nil, result.Error
	}
	return task, nil
}

// CompleteTask marks a task complete and, for recurring tasks, advances the due date.
// completionDate must be in YYYY-MM-DD format.
func CompleteTask(db *gorm.DB, id uint, completionDate string) (*models.Task, error) {
	task, err := GetTask(db, id)
	if err != nil {
		return nil, err
	}

	task.LastCompletedAt = &completionDate

	if !task.IsRecurring {
		task.Checked = true
	} else {
		// Determine the base date for advancing the schedule
		baseDate := completionDate
		if task.RecurrenceMode == "due_date" && task.DueDate != nil && *task.DueDate != "" {
			baseDate = *task.DueDate
		}

		nextDue, err := advanceDate(baseDate, task.RecurrenceUnit, task.RecurrenceInterval)
		if err != nil {
			return nil, fmt.Errorf("error computing next due date: %w", err)
		}
		task.DueDate = &nextDue
	}

	result := db.Save(task)
	if result.Error != nil {
		return nil, result.Error
	}
	return task, nil
}

// UncompleteTask marks a non-recurring task as incomplete. No-op for recurring tasks.
func UncompleteTask(db *gorm.DB, id uint) (*models.Task, error) {
	task, err := GetTask(db, id)
	if err != nil {
		return nil, err
	}

	if !task.IsRecurring {
		task.Checked = false
		result := db.Save(task)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return task, nil
}

// DeleteTask deletes a task by ID.
func DeleteTask(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.Task{})
	return result.Error
}

// advanceDate adds the given interval in the given unit to a YYYY-MM-DD date string.
func advanceDate(base, unit string, interval int) (string, error) {
	t, err := time.Parse("2006-01-02", base)
	if err != nil {
		return "", err
	}

	var next time.Time
	switch unit {
	case "days":
		next = t.AddDate(0, 0, interval)
	case "weeks":
		next = t.AddDate(0, 0, interval*7)
	case "months":
		next = t.AddDate(0, interval, 0)
	case "years":
		next = t.AddDate(interval, 0, 0)
	default:
		next = t.AddDate(0, interval, 0)
	}

	return next.Format("2006-01-02"), nil
}

// MigrateTodosToTasks copies any rows from the todos table that have not yet
// been migrated into tasks. It uses a migration-tracking table
// (todo_task_migrations) to record which todo IDs have already been converted,
// so it is safe to call on every startup.
func MigrateTodosToTasks(db *gorm.DB) error {
	// Ensure the tracking table exists.
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS todo_task_migrations (todo_id INTEGER PRIMARY KEY)`).Error; err != nil {
		return fmt.Errorf("create migration tracking table: %w", err)
	}

	// Load all todos that haven't been migrated yet.
	var todos []models.Todo
	if err := db.Raw(`
		SELECT * FROM todos
		WHERE deleted_at IS NULL
		  AND id NOT IN (SELECT todo_id FROM todo_task_migrations)
	`).Scan(&todos).Error; err != nil {
		return fmt.Errorf("query pending todos: %w", err)
	}

	for _, t := range todos {
		task := &models.Task{
			Label:   t.Label,
			Checked: t.Checked,
			UserID:  t.UserID,
		}
		if t.ApplianceID != nil {
			aid := *t.ApplianceID
			task.ApplianceID = &aid
		}
		if t.SpaceType != nil {
			st := *t.SpaceType
			task.SpaceType = &st
		}

		if err := db.Create(task).Error; err != nil {
			fmt.Printf("MigrateTodosToTasks: skipping todo %d: %v\n", t.ID, err)
			continue
		}

		// Record as migrated.
		if err := db.Exec(`INSERT INTO todo_task_migrations (todo_id) VALUES (?)`, t.ID).Error; err != nil {
			fmt.Printf("MigrateTodosToTasks: failed to record migration for todo %d: %v\n", t.ID, err)
		}
	}

	if len(todos) > 0 {
		fmt.Printf("MigrateTodosToTasks: migrated %d todo(s) to tasks\n", len(todos))
	}
	return nil
}
