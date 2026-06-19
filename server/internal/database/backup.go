package database

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// exportDataToJson fetches all data for registered models and marshals it into a JSON byte array.
func ExportDataToJson(db *gorm.DB) ([]byte, error) {
	data := make(map[string]interface{})

	// List of models to back up. This should match `MigrateGorm` in gorm.go
	modelsToBackup := []interface{}{
		&models.Todo{},
		&models.Appliance{},
		&models.Maintenance{},
		&models.Repair{},
		&models.SavedFile{},
		&models.Note{},
		&models.Task{},
	}

	for _, model := range modelsToBackup {
		modelType := reflect.TypeOf(model).Elem()
		modelName := modelType.Name()

		// Create a slice to hold records of the current model type
		sliceType := reflect.SliceOf(modelType)
		records := reflect.New(sliceType).Interface()

		if err := db.Find(records).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch %s records: %w", modelName, err)
		}
		data[modelName] = reflect.ValueOf(records).Elem().Interface()
	}

	backupContent := map[string]interface{}{
		"version": "1.0",
		"data":    data,
	}

	jsonData, err := json.MarshalIndent(backupContent, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup data to JSON: %w", err)
	}

	return jsonData, nil
}

// ponytail: No generic GORM GetAll function. Create a specific backup file.
