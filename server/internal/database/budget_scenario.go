package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetBudgetScenarios gets all budget scenarios
func GetBudgetScenarios(db *gorm.DB) ([]models.BudgetScenario, error) {
	var scenarios []models.BudgetScenario
	result := db.Find(&scenarios)
	if result.Error != nil {
		return []models.BudgetScenario{}, result.Error
	}

	return scenarios, nil
}

// AddBudgetScenario creates a new budget scenario
func AddBudgetScenario(db *gorm.DB, scenario *models.BudgetScenario) (*models.BudgetScenario, error) {
	result := db.Create(scenario)
	if result.Error != nil {
		return nil, result.Error
	}

	return scenario, nil
}

// UpdateBudgetScenario updates a budget scenario
func UpdateBudgetScenario(db *gorm.DB, scenario *models.BudgetScenario) (*models.BudgetScenario, error) {
	result := db.Save(scenario)
	if result.Error != nil {
		return nil, result.Error
	}

	return scenario, nil
}

// GetBudgetScenario gets a budget scenario by ID
func GetBudgetScenario(db *gorm.DB, id uint) (*models.BudgetScenario, error) {
	var scenario models.BudgetScenario
	result := db.Where("id = ?", id).First(&scenario)
	if result.Error != nil {
		return nil, result.Error
	}

	return &scenario, nil
}

// DeleteBudgetScenario deletes a budget scenario by ID
func DeleteBudgetScenario(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.BudgetScenario{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
