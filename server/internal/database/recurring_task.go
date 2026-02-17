package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetRecurringTasks gets all recurring tasks
func GetRecurringTasks(db *gorm.DB) ([]models.RecurringTask, error) {
	var tasks []models.RecurringTask
	result := db.Preload("Category").Preload("Appliance").Find(&tasks)
	if result.Error != nil {
		return []models.RecurringTask{}, result.Error
	}

	return tasks, nil
}

// AddRecurringTask creates a new recurring task
func AddRecurringTask(db *gorm.DB, task *models.RecurringTask) (*models.RecurringTask, error) {
	result := db.Create(task)
	if result.Error != nil {
		return nil, result.Error
	}

	return task, nil
}

// UpdateRecurringTask updates a recurring task
func UpdateRecurringTask(db *gorm.DB, task *models.RecurringTask) (*models.RecurringTask, error) {
	result := db.Save(task)
	if result.Error != nil {
		return nil, result.Error
	}

	return task, nil
}

// GetRecurringTask gets a recurring task by ID
func GetRecurringTask(db *gorm.DB, id uint) (*models.RecurringTask, error) {
	var task models.RecurringTask
	result := db.Preload("Category").Preload("Appliance").Where("id = ?", id).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}

	return &task, nil
}

// DeleteRecurringTask deletes a recurring task by ID
func DeleteRecurringTask(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.RecurringTask{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
