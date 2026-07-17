package database

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// ConvertLegacyDB opens a legacy SQLite backup file (from the old backup format)
// and reads all entity tables into a BackupPayload, compatible with ImportFromJSON.
// Tables that don't exist in the old DB (e.g. "tasks" in pre-migration backups)
// are skipped gracefully.
func ConvertLegacyDB(dbPath string) (*models.BackupPayload, error) {
	oldDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open legacy db: %w", err)
	}

	sqlDB, err := oldDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	payload := &models.BackupPayload{
		Version:      BackupVersion,
		ExportedAt:   time.Now().UTC(),
		DatabaseType: dialectSQLite,
	}

	migrator := oldDB.Migrator()

	if migrator.HasTable("appliances") {
		if err := oldDB.Find(&payload.Entities.Appliances).Error; err != nil {
			return nil, fmt.Errorf("read appliances: %w", err)
		}
	}

	if migrator.HasTable("todos") {
		if err := oldDB.Find(&payload.Entities.Todos).Error; err != nil {
			return nil, fmt.Errorf("read todos: %w", err)
		}
	}

	if migrator.HasTable("tasks") {
		if err := oldDB.Find(&payload.Entities.Tasks).Error; err != nil {
			return nil, fmt.Errorf("read tasks: %w", err)
		}
	}

	if migrator.HasTable("maintenances") {
		if err := oldDB.Find(&payload.Entities.Maintenance).Error; err != nil {
			return nil, fmt.Errorf("read maintenances: %w", err)
		}
	}

	if migrator.HasTable("repairs") {
		if err := oldDB.Find(&payload.Entities.Repairs).Error; err != nil {
			return nil, fmt.Errorf("read repairs: %w", err)
		}
	}

	if migrator.HasTable("saved_files") {
		if err := oldDB.Find(&payload.Entities.SavedFiles).Error; err != nil {
			return nil, fmt.Errorf("read saved_files: %w", err)
		}
	}

	if migrator.HasTable("notes") {
		if err := oldDB.Find(&payload.Entities.Notes).Error; err != nil {
			return nil, fmt.Errorf("read notes: %w", err)
		}
	}

	return payload, nil
}
