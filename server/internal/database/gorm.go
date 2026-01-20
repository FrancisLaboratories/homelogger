package database

import (
	"os"
	"path/filepath"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConnectGorm connects to the database
func ConnectGorm() (*gorm.DB, error) {
	dbPath := "./data/db/homelogger.db"
	dir := filepath.Dir(dbPath)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// If the DB file doesn't exist, create an empty file so sqlite can open it
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// MigrateGorm migrates the database
func MigrateGorm(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Todo{}, &models.Appliance{}, &models.Maintenance{}, &models.Repair{}, &models.SavedFile{}, &models.Note{})
	if err != nil {
		return err
	}

	return nil
}
