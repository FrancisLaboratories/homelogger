package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetAppliances gets all appliances
func GetAppliances(db *gorm.DB) ([]models.Appliance, error) {
	var appliances []models.Appliance
	result := db.Find(&appliances)
	if result.Error != nil {
		return []models.Appliance{}, result.Error
	}

	return appliances, nil
}

// AddAppliance creates a new appliance
func AddAppliance(db *gorm.DB, appliance *models.Appliance) (*models.Appliance, error) {
	result := db.Create(appliance)
	if result.Error != nil {
		return nil, result.Error
	}

	return appliance, nil
}

// UpdateAppliance updates an appliance, saving the changes to an existing appliance by its ID
func UpdateAppliance(db *gorm.DB, appliance *models.Appliance) (*models.Appliance, error) {
	result := db.Save(appliance)
	if result.Error != nil {
		return nil, result.Error
	}

	return appliance, nil
}

// GetAppliance gets an appliance by ID
func GetAppliance(db *gorm.DB, id uint) (*models.Appliance, error) {
	var appliance models.Appliance
	result := db.Where("id = ?", id).First(&appliance)
	if result.Error != nil {
		return nil, result.Error
	}

	return &appliance, nil
}

// DeleteAppliance deletes an appliance by ID
func DeleteAppliance(db *gorm.DB, id uint) error {
	// Delete files attached directly to the appliance
	if err := DeleteFilesByAppliance(db, id); err != nil {
		return err
	}

	// For each maintenance linked to this appliance, delete its files
	var maintenances []models.Maintenance
	if r := db.Select("id").Where("appliance_id = ?", id).Find(&maintenances); r.Error != nil {
		return r.Error
	}
	for _, m := range maintenances {
		if err := DeleteFilesByMaintenance(db, m.ID); err != nil {
			return err
		}
	}

	// Delete maintenance rows for this appliance
	if err := db.Where("appliance_id = ?", id).Delete(&models.Maintenance{}).Error; err != nil {
		return err
	}

	// For each repair linked to this appliance, delete its files
	var repairs []models.Repair
	if r := db.Select("id").Where("appliance_id = ?", id).Find(&repairs); r.Error != nil {
		return r.Error
	}
	for _, rp := range repairs {
		if err := DeleteFilesByRepair(db, rp.ID); err != nil {
			return err
		}
	}

	// Delete repair rows for this appliance
	if err := db.Where("appliance_id = ?", id).Delete(&models.Repair{}).Error; err != nil {
		return err
	}

	var improvements []models.Improvement
	if err := db.Select("id").Where("appliance_id = ?", id).Find(&improvements).Error; err != nil {
		return err
	}
	for _, improvement := range improvements {
		if err := DeleteFilesByImprovement(db, improvement.ID); err != nil {
			return err
		}
	}
	if err := db.Where("appliance_id = ?", id).Delete(&models.Improvement{}).Error; err != nil {
		return err
	}

	// Finally delete the appliance
	result := db.Where("id = ?", id).Delete(&models.Appliance{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
