package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetBudgetCategories gets all budget categories
func GetBudgetCategories(db *gorm.DB) ([]models.BudgetCategory, error) {
	var categories []models.BudgetCategory
	result := db.Find(&categories)
	if result.Error != nil {
		return []models.BudgetCategory{}, result.Error
	}

	return categories, nil
}

// AddBudgetCategory creates a new budget category
func AddBudgetCategory(db *gorm.DB, category *models.BudgetCategory) (*models.BudgetCategory, error) {
	result := db.Create(category)
	if result.Error != nil {
		return nil, result.Error
	}

	return category, nil
}

// UpdateBudgetCategory updates a budget category
func UpdateBudgetCategory(db *gorm.DB, category *models.BudgetCategory) (*models.BudgetCategory, error) {
	result := db.Save(category)
	if result.Error != nil {
		return nil, result.Error
	}

	return category, nil
}

// GetBudgetCategory gets a budget category by ID
func GetBudgetCategory(db *gorm.DB, id uint) (*models.BudgetCategory, error) {
	var category models.BudgetCategory
	result := db.Where("id = ?", id).First(&category)
	if result.Error != nil {
		return nil, result.Error
	}

	return &category, nil
}

// DeleteBudgetCategory deletes a budget category by ID
func DeleteBudgetCategory(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.BudgetCategory{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
