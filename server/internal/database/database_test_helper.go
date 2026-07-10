package database

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func testDialect() string {
    d := strings.ToLower(strings.TrimSpace(os.Getenv("TEST_DB_DIALECT")))
    if d == "" {
        return "sqlite"
    }
    if d == "postgresql" {
        return "postgres"
    }
    return d
}

func postgresTestDSN() string {
    if dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); dsn != "" {
        return dsn
    }
    if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
        return dsn
    }

    host := strings.TrimSpace(os.Getenv("TEST_DB_HOST"))
    if host == "" {
        host = strings.TrimSpace(os.Getenv("DB_HOST"))
    }
    if host == "" {
        host = "localhost"
    }

    port := strings.TrimSpace(os.Getenv("TEST_DB_PORT"))
    if port == "" {
        port = strings.TrimSpace(os.Getenv("DB_PORT"))
    }
    if port == "" {
        port = "5432"
    }

    user := strings.TrimSpace(os.Getenv("TEST_DB_USER"))
    if user == "" {
        user = strings.TrimSpace(os.Getenv("DB_USER"))
    }

    password := os.Getenv("TEST_DB_PASSWORD")
    if password == "" {
        password = os.Getenv("DB_PASSWORD")
    }

    dbName := strings.TrimSpace(os.Getenv("TEST_DB_NAME"))
    if dbName == "" {
        dbName = strings.TrimSpace(os.Getenv("DB_NAME"))
    }
    if dbName == "" {
        return ""
    }

    sslmode := strings.TrimSpace(os.Getenv("TEST_DB_SSLMODE"))
    if sslmode == "" {
        sslmode = strings.TrimSpace(os.Getenv("DB_SSLMODE"))
    }
    if sslmode == "" {
        sslmode = "disable"
    }

    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslmode)
}

func resetPostgresTestSchema(db *gorm.DB) error {
    tables := []string{
        "todo_task_migrations",
        "tasks",
        "notes",
        "saved_files",
        "repairs",
        "maintenances",
        "appliances",
        "todos",
    }

    for _, table := range tables {
        if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
            return err
        }
    }

    return nil
}

// TestDB creates a clean test DB per test run and migrates models.
// Default dialect is sqlite. Set TEST_DB_DIALECT=postgres to run against PostgreSQL.
func TestDB(t *testing.T) *gorm.DB {
    t.Helper()

    dialect := testDialect()
    var db *gorm.DB
    var err error

    switch dialect {
    case "postgres":
        dsn := postgresTestDSN()
        if dsn == "" {
            t.Skip("TEST_DB_DIALECT=postgres requires TEST_DATABASE_URL or TEST_DB_NAME/DB_NAME configuration")
        }
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err != nil {
            t.Fatalf("failed to open postgres test db: %v", err)
        }
        if err := resetPostgresTestSchema(db); err != nil {
            t.Fatalf("failed to reset postgres test schema: %v", err)
        }
    default:
        // Use a unique DSN per test to avoid sharing state between tests.
        dsn := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
        db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
        if err != nil {
            t.Fatalf("failed to open in-memory db: %v", err)
        }
    }

    if err := MigrateGorm(db); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }

    if err := MigrateTodosToTasks(db); err != nil {
        t.Fatalf("failed to run todo->task migration: %v", err)
    }

    if sqlDB, err := db.DB(); err == nil {
        t.Cleanup(func() { _ = sqlDB.Close() })
    }

    return db
}
