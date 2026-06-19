package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	defaultSQLiteDBPath   = "./data/db/homelogger.db"
	defaultEngineTXTName  = "db_engine.txt"
	defaultEngineDir      = "./data/db"
	envEngineMetadataPath = "DB_ENGINE_METADATA_PATH"
)

type DatabaseConfig struct {
	Driver string
	DSN    string
}

// GetDatabaseConfig returns the database connection config and enforces engine pinning.
func GetDatabaseConfig() (*DatabaseConfig, error) {
	return parseDatabaseConfig()
}

func parseDatabaseConfig() (*DatabaseConfig, error) {
	dbType := strings.TrimSpace(strings.ToLower(os.Getenv("DB_TYPE")))
	if dbType == "" {
		dbType = "sqlite"
	}
	switch dbType {
	case "sqlite", "postgres", "mariadb", "mysql":
		// allow mysql alias for mariadb
		if dbType == "mysql" {
			dbType = "mariadb"
		}
	default:
		return nil, fmt.Errorf("unsupported DB_TYPE %q; valid values are sqlite, postgres, mariadb", dbType)
	}

	if os.Getenv("DEMO_MODE") == "true" || os.Getenv("DEMO_MODE") == "1" {
		if dbType != "sqlite" {
			return nil, fmt.Errorf("DEMO_MODE requires sqlite database engine, but DB_TYPE is %q", dbType)
		}
	}

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbType == "sqlite" {
		if demoPath := strings.TrimSpace(os.Getenv("DEMO_DB_PATH")); demoPath != "" {
			dsn = demoPath
		}
		if dsn == "" {
			dsn = defaultSQLiteDBPath
		}
	} else {
		if dsn == "" {
			return nil, fmt.Errorf("DATABASE_URL is required for DB_TYPE=%s", dbType)
		}
	}

	engineFilePath := getEngineMetadataPath(dbType, dsn)
	if err := ensureEngineMetadata(dbType, engineFilePath); err != nil {
		return nil, err
	}

	return &DatabaseConfig{Driver: dbType, DSN: dsn}, nil
}

func getEngineMetadataPath(dbType, dsn string) string {
	if override := strings.TrimSpace(os.Getenv(envEngineMetadataPath)); override != "" {
		return override
	}
	if dbType == "sqlite" {
		if isSQLiteFilePath(dsn) {
			return filepath.Join(filepath.Dir(dsn), defaultEngineTXTName)
		}
	}
	return filepath.Join(defaultEngineDir, defaultEngineTXTName)
}

func isSQLiteFilePath(dsn string) bool {
	return !strings.HasPrefix(dsn, "file:")
}

func ensureEngineMetadata(driver, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return os.WriteFile(filePath, []byte(driver), 0644)
	} else if err != nil {
		return err
	}

	current, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	persisted := strings.TrimSpace(strings.ToLower(string(current)))
	if persisted == "" {
		return os.WriteFile(filePath, []byte(driver), 0644)
	}
	if persisted != driver {
		return fmt.Errorf("database engine is locked to %q; current DB_TYPE=%q does not match", persisted, driver)
	}
	return nil
}

// ConnectGorm connects to the database.
func ConnectGorm() (*gorm.DB, error) {
	config, err := GetDatabaseConfig()
	if err != nil {
		return nil, err
	}

	var dialector gorm.Dialector
	switch config.Driver {
	case "sqlite":
		dsn := config.DSN
		if isSQLiteFilePath(dsn) {
			dir := filepath.Dir(dsn)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			if _, err := os.Stat(dsn); os.IsNotExist(err) {
				f, err := os.Create(dsn)
				if err != nil {
					return nil, err
				}
				_ = f.Close()
			}
		}
		dialector = sqlite.Open(dsn)
	case "postgres":
		dialector = postgres.Open(config.DSN)
	case "mariadb":
		dialector = mysql.Open(config.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", config.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// MigrateGorm migrates the database
func MigrateGorm(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Todo{}, &models.Appliance{}, &models.Maintenance{}, &models.Repair{}, &models.SavedFile{}, &models.Note{}, &models.Task{})
	if err != nil {
		return err
	}

	// Drop legacy associated_id column from saved_files if it still exists.
	// This column was removed from the model but AutoMigrate never drops columns.
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('saved_files') WHERE name = 'associated_id'").Scan(&count)
	if count > 0 {
		if err := db.Exec("ALTER TABLE saved_files DROP COLUMN associated_id").Error; err != nil {
			return err
		}
	}

	return nil
}
