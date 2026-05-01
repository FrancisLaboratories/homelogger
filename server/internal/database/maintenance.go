package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetMaintenances gets maintenance records filtered by applianceId, referenceType, and spaceType
func GetMaintenances(db *gorm.DB, applianceId uint, referenceType, spaceType string) ([]models.Maintenance, error) {
	var maintenances []models.Maintenance
	if referenceType == "Space" {
		result := db.Where("reference_type = ? AND space_type = ?", referenceType, spaceType).Find(&maintenances)
		if result.Error != nil {
			return []models.Maintenance{}, result.Error
		}

		return maintenances, nil
	} else {
		result := db.Where("appliance_id = ? AND reference_type = ?", applianceId, referenceType).Find(&maintenances)
		if result.Error != nil {
			return []models.Maintenance{}, result.Error
		}

		return maintenances, nil
	}

}

// AddMaintenance creates a new maintenance record
func AddMaintenance(db *gorm.DB, maintenance *models.Maintenance) (*models.Maintenance, error) {
	result := db.Create(maintenance)
	if result.Error != nil {
		return nil, result.Error
	}

	return maintenance, nil
}

// GetMaintenance gets a maintenance record by ID
func GetMaintenance(db *gorm.DB, id uint) (*models.Maintenance, error) {
	var maintenance models.Maintenance
	result := db.Where("id = ?", id).First(&maintenance)
	if result.Error != nil {
		return nil, result.Error
	}

	return &maintenance, nil
}

// UpdateMaintenance updates the editable fields of an existing maintenance record.
func UpdateMaintenance(db *gorm.DB, id uint, description, date string, cost float64, notes string) (*models.Maintenance, error) {
	maintenance, err := GetMaintenance(db, id)
	if err != nil {
		return nil, err
	}
	maintenance.Description = description
	maintenance.Date = date
	maintenance.Cost = cost
	maintenance.Notes = notes
	result := db.Save(maintenance)
	if result.Error != nil {
		return nil, result.Error
	}
	return maintenance, nil
}

// DeleteMaintenance deletes a maintenance record by ID
func DeleteMaintenance(db *gorm.DB, id uint) error {
	// Remove associated files (disk + DB)
	if err := DeleteFilesByMaintenance(db, id); err != nil {
		return err
	}

	result := db.Where("id = ?", id).Delete(&models.Maintenance{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
