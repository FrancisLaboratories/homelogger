package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetPlannedCosts gets all planned costs (optionally filtered by scenarioId)
func GetPlannedCosts(db *gorm.DB, scenarioId uint) ([]models.PlannedCost, error) {
	var costs []models.PlannedCost
	query := db.Preload("Scenario").Preload("Category")
	if scenarioId != 0 {
		query = query.Where("scenario_id = ?", scenarioId)
	}
	result := query.Find(&costs)
	if result.Error != nil {
		return []models.PlannedCost{}, result.Error
	}

	return costs, nil
}

// AddPlannedCost creates a new planned cost
func AddPlannedCost(db *gorm.DB, cost *models.PlannedCost) (*models.PlannedCost, error) {
	result := db.Create(cost)
	if result.Error != nil {
		return nil, result.Error
	}

	return cost, nil
}

// UpdatePlannedCost updates a planned cost
func UpdatePlannedCost(db *gorm.DB, cost *models.PlannedCost) (*models.PlannedCost, error) {
	result := db.Save(cost)
	if result.Error != nil {
		return nil, result.Error
	}

	return cost, nil
}

// GetPlannedCost gets a planned cost by ID
func GetPlannedCost(db *gorm.DB, id uint) (*models.PlannedCost, error) {
	var cost models.PlannedCost
	result := db.Preload("Scenario").Preload("Category").Where("id = ?", id).First(&cost)
	if result.Error != nil {
		return nil, result.Error
	}

	return &cost, nil
}

// DeletePlannedCost deletes a planned cost by ID
func DeletePlannedCost(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.PlannedCost{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
