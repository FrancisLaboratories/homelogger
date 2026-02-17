package database

import (
	"errors"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

func DefaultSettings() models.Settings {
	return models.Settings{
		Locale:            "en-US",
		Language:          "en",
		Currency:          "USD",
		TimeZone:          "UTC",
		MeasurementSystem: "metric",
		WeekStart:         0,
		DateFormat:        "YYYY-MM-DD",
		NumberingSystem:   "latn",
	}
}

	// EnsureSettings guarantees a single settings row exists.
	func EnsureSettings(db *gorm.DB) (*models.Settings, error) {
		var settings models.Settings
		result := db.Limit(1).Find(&settings)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected > 0 {
			return &settings, nil
		}

		defaultSettings := DefaultSettings()
	if err := db.Create(&defaultSettings).Error; err != nil {
		return nil, err
	}

	return &defaultSettings, nil
}

// GetSettings returns the single settings record, creating defaults if missing.
func GetSettings(db *gorm.DB) (*models.Settings, error) {
	return EnsureSettings(db)
}

// UpdateSettings updates and persists the settings record.
func UpdateSettings(db *gorm.DB, settings *models.Settings) (*models.Settings, error) {
	if err := db.Save(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}
