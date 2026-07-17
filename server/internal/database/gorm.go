package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	dialectSQLite   = "sqlite"
	dialectPostgres = "postgres"

	// Exported aliases for use in tests and external packages.
	DialectSQLite   = dialectSQLite
	DialectPostgres = dialectPostgres

	defaultDialectLockPath = "./data/db/.db_dialect"
)

type dialectSelection struct {
	dialect      string
	lockPath     string
	lockExists   bool
	lockedDialect string
	force        bool
}

func normalizeDialect(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", dialectSQLite:
		return dialectSQLite
	case "postgres", "postgresql":
		return dialectPostgres
	default:
		return ""
	}
}

func dialectLockPathFromEnv() string {
	if lockPath := strings.TrimSpace(os.Getenv("DB_DIALECT_LOCK_PATH")); lockPath != "" {
		return lockPath
	}
	return defaultDialectLockPath
}

func forceDialectChangeEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("FORCE_DB_DIALECT_CHANGE")))
	return v == "1" || v == "true"
}

func loadLockedDialect(lockPath string) (string, bool, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	locked := normalizeDialect(string(data))
	if locked == "" {
		return "", false, fmt.Errorf("invalid value in DB dialect lock file %q", lockPath)
	}

	return locked, true, nil
}

func saveLockedDialect(lockPath, dialect string) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(lockPath, []byte(dialect), 0644)
}

func selectDialect() (dialectSelection, error) {
	requestedRaw := os.Getenv("DB_DIALECT")
	requested := normalizeDialect(requestedRaw)
	if requested == "" {
		return dialectSelection{}, fmt.Errorf("invalid DB_DIALECT value: %q", requestedRaw)
	}

	selection := dialectSelection{
		dialect:  requested,
		lockPath: dialectLockPathFromEnv(),
		force:    forceDialectChangeEnabled(),
	}

	locked, exists, err := loadLockedDialect(selection.lockPath)
	if err != nil {
		return dialectSelection{}, err
	}

	selection.lockExists = exists
	selection.lockedDialect = locked

	if !exists {
		return selection, nil
	}

	if locked != requested && !selection.force {
		return dialectSelection{}, fmt.Errorf(
			"this instance is configured for %q and cannot be changed after first run; Please set DB_DIALECT to %q",
			locked,
			locked,
		)
	}

	return selection, nil
}

func buildPostgresDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	if port == "" {
		port = "5432"
	}

	user := strings.TrimSpace(os.Getenv("DB_USER"))
	dbname := strings.TrimSpace(os.Getenv("DB_NAME"))
	password := os.Getenv("DB_PASSWORD")
	sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
}

func sqlitePathFromEnv() string {
	if dbPath := strings.TrimSpace(os.Getenv("DEMO_DB_PATH")); dbPath != "" {
		return dbPath
	}
	if dbPath := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbPath != "" {
		return dbPath
	}
	return "./data/db/homelogger.db"
}

func ensureSQLiteFile(dbPath string) error {
	dir := filepath.Dir(dbPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.Create(dbPath)
		if err != nil {
			return err
		}
		_ = f.Close()
	}

	return nil
}

func columnExists(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	switch db.Dialector.Name() {
	case dialectSQLite:
		if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", tableName, columnName).Scan(&count).Error; err != nil {
			return false, err
		}
	case dialectPostgres:
		if err := db.Raw(`
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = ?
			  AND column_name = ?
		`, tableName, columnName).Scan(&count).Error; err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("unsupported database dialect: %s", db.Dialector.Name())
	}

	return count > 0, nil
}

// ConnectGorm connects to the database
func ConnectGorm() (*gorm.DB, error) {
	selection, err := selectDialect()
	if err != nil {
		return nil, err
	}

	dialect := selection.dialect
	var db *gorm.DB

	switch dialect {
	case dialectPostgres:
		dsn := buildPostgresDSN()
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case dialectSQLite:
		fallthrough
	default:
		dbPath := sqlitePathFromEnv()
		if err := ensureSQLiteFile(dbPath); err != nil {
			return nil, err
		}
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}

	if err != nil {
		return nil, err
	}

	shouldPersistLock := !selection.lockExists || (selection.force && selection.lockedDialect != dialect)
	if shouldPersistLock {
		if err := saveLockedDialect(selection.lockPath, dialect); err != nil {
			return nil, err
		}
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
	exists, err := columnExists(db, "saved_files", "associated_id")
	if err != nil {
		return err
	}
	if exists {
		if err := db.Exec("ALTER TABLE saved_files DROP COLUMN associated_id").Error; err != nil {
			return err
		}
	}

	return nil
}
