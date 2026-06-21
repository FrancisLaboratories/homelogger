package database

import (
	"fmt"
	"time"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

const BackupVersion = "1.0"

// ExportToJSON fetches all data and returns a typed BackupPayload.
// Works with any GORM dialect — no raw SQL, no dialect-specific logic.
func ExportToJSON(db *gorm.DB, dbType string) (*models.BackupPayload, error) {
	payload := &models.BackupPayload{
		Version:      BackupVersion,
		ExportedAt:   time.Now().UTC(),
		DatabaseType: dbType,
	}

	if err := db.Find(&payload.Entities.Appliances).Error; err != nil {
		return nil, fmt.Errorf("fetch Appliance: %w", err)
	}
	if err := db.Find(&payload.Entities.Tasks).Error; err != nil {
		return nil, fmt.Errorf("fetch Task: %w", err)
	}
	if err := db.Find(&payload.Entities.Maintenance).Error; err != nil {
		return nil, fmt.Errorf("fetch Maintenance: %w", err)
	}
	if err := db.Find(&payload.Entities.Repairs).Error; err != nil {
		return nil, fmt.Errorf("fetch Repair: %w", err)
	}
	if err := db.Find(&payload.Entities.SavedFiles).Error; err != nil {
		return nil, fmt.Errorf("fetch SavedFile: %w", err)
	}
	if err := db.Find(&payload.Entities.Notes).Error; err != nil {
		return nil, fmt.Errorf("fetch Note: %w", err)
	}
	if err := db.Find(&payload.Entities.Todos).Error; err != nil {
		return nil, fmt.Errorf("fetch Todo: %w", err)
	}

	return payload, nil
}
