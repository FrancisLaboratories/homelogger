package database

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"path/filepath"
	"os"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// ImportDataFromJson wipes existing data and imports new data from a JSON file.
func ImportDataFromJson(db *gorm.DB, jsonFilePath string) error {
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return fmt.Errorf("failed to read data.json: %w", err)
	}

	var backupContent struct {
		Version string                            `json:"version"`
		Data    map[string][]map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(jsonData, &backupContent); err != nil {
		return fmt.Errorf("failed to unmarshal data.json: %w", err)
	}

	// Models in reverse order for deletion to respect foreign keys
	modelsToDelete := []interface{}{
		&models.Task{},
		&models.Note{},
		&models.SavedFile{},
		&models.Repair{},
		&models.Maintenance{},
		&models.Appliance{},
		&models.Todo{},
	}

	// Wipe existing data
	for _, model := range modelsToDelete {
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
			return fmt.Errorf("failed to wipe %s data: %w", reflect.TypeOf(model).Elem().Name(), err)
		}
	}

	// Models in dependency order for creation
	modelsToCreate := []interface{}{
		&models.Appliance{},
		&models.Todo{},
		&models.Maintenance{},
		&models.Repair{},
		&models.SavedFile{},
		&models.Note{},
		&models.Task{},
	}

	// Import new data
	for _, model := range modelsToCreate {
		modelType := reflect.TypeOf(model).Elem()
		modelName := modelType.Name()

		if records, ok := backupContent.Data[modelName]; ok {
			for _, record := range records {
				newRecord := reflect.New(modelType).Interface()
				
				// Marshal then unmarshal to handle type conversions and GORM tags
				recordBytes, err := json.Marshal(record)
				if err != nil {
					return fmt.Errorf("failed to marshal record to bytes for %s: %w", modelName, err)
				}
				if err := json.Unmarshal(recordBytes, newRecord); err != nil {
					return fmt.Errorf("failed to unmarshal record into %s struct: %w", modelName, err)
				}

				// ponytail: Explicitly set CreatedAt, UpdatedAt, DeletedAt to preserve original timestamps
				// This requires GORM to not ignore these fields on Create.
				// By default, GORM will update CreatedAt/UpdatedAt, but we want to preserve them.
				// This is generally handled by using `db.Omit("CreatedAt", "UpdatedAt").Create(...)` or by
				// setting the fields directly on the struct and GORM should honor them if not zero value.
				// For deleted at, if it's not null, GORM will treat it as a soft-deleted record.
				
				if err := db.Create(newRecord).Error; err != nil {
					return fmt.Errorf("failed to create %s record with ID %v: %w", modelName, reflect.ValueOf(newRecord).Elem().FieldByName("ID").Uint(), err)
				}
			}
		}
	}

	return nil
}

// ImportUploads copies extracted uploads to the application's uploads directory.
func ImportUploads(extractedUploadsPath string) error {
	appUploadsRoot := "./data/uploads"

	// ponytail: Wipe existing uploads directory first
	if err := os.RemoveAll(appUploadsRoot); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing uploads directory: %w", err)
	}
	if err := os.MkdirAll(appUploadsRoot, 0755); err != nil {
		return fmt.Errorf("failed to create application uploads directory: %w", err)
	}

	if extractedUploadsPath == "" {
		return nil // No uploads to import
	}

	// Copy all files from extractedUploadsPath to appUploadsRoot
	err := filepath.Walk(extractedUploadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(extractedUploadsPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		destPath := filepath.Join(appUploadsRoot, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(destPath), err)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", path, err)
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", path, destPath, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to import uploaded files: %w", err)
	}

	return nil
}
