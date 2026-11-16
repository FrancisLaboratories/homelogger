package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

func GetImprovements(db *gorm.DB, applianceID uint, referenceType, spaceType string) ([]models.Improvement, error) {
	var improvements []models.Improvement
	query := db
	if referenceType == "Space" {
		query = query.Where("reference_type = ? AND space_type = ?", referenceType, spaceType)
	} else {
		query = query.Where("appliance_id = ? AND reference_type = ?", applianceID, referenceType)
	}
	if err := query.Find(&improvements).Error; err != nil {
		return []models.Improvement{}, err
	}
	return improvements, nil
}

func AddImprovement(db *gorm.DB, improvement *models.Improvement) (*models.Improvement, error) {
	if improvement.ApplianceID != nil && *improvement.ApplianceID == 0 {
		improvement.ApplianceID = nil
	}
	if err := db.Create(improvement).Error; err != nil {
		return nil, err
	}
	return improvement, nil
}

func GetImprovement(db *gorm.DB, id uint) (*models.Improvement, error) {
	var improvement models.Improvement
	if err := db.Where("id = ?", id).First(&improvement).Error; err != nil {
		return nil, err
	}
	return &improvement, nil
}

func UpdateImprovement(db *gorm.DB, id uint, description, date string, cost float64, notes string) (*models.Improvement, error) {
	improvement, err := GetImprovement(db, id)
	if err != nil {
		return nil, err
	}
	improvement.Description = description
	improvement.Date = date
	improvement.Cost = cost
	improvement.Notes = notes
	if err := db.Save(improvement).Error; err != nil {
		return nil, err
	}
	return improvement, nil
}

func DeleteImprovement(db *gorm.DB, id uint) error {
	if err := DeleteFilesByImprovement(db, id); err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(&models.Improvement{}).Error
}
