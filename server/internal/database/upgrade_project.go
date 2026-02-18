package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetUpgradeProjects gets all upgrade projects
func GetUpgradeProjects(db *gorm.DB) ([]models.UpgradeProject, error) {
	var projects []models.UpgradeProject
	result := db.Preload("Category").Find(&projects)
	if result.Error != nil {
		return []models.UpgradeProject{}, result.Error
	}

	return projects, nil
}

// AddUpgradeProject creates a new upgrade project
func AddUpgradeProject(db *gorm.DB, project *models.UpgradeProject) (*models.UpgradeProject, error) {
	result := db.Create(project)
	if result.Error != nil {
		return nil, result.Error
	}

	return project, nil
}

// UpdateUpgradeProject updates an upgrade project
func UpdateUpgradeProject(db *gorm.DB, project *models.UpgradeProject) (*models.UpgradeProject, error) {
	result := db.Save(project)
	if result.Error != nil {
		return nil, result.Error
	}

	return project, nil
}

// GetUpgradeProject gets an upgrade project by ID
func GetUpgradeProject(db *gorm.DB, id uint) (*models.UpgradeProject, error) {
	var project models.UpgradeProject
	result := db.Preload("Category").Where("id = ?", id).First(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}

// DeleteUpgradeProject deletes an upgrade project by ID
func DeleteUpgradeProject(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.UpgradeProject{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
