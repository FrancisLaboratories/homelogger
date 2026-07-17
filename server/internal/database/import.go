package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// tableDropOrder lists tables in reverse FK dependency order for safe drops.
// note: hard-coded list mirrors MigrateGorm — update both together.
var tableDropOrder = []string{
	"tasks",
	"notes",
	"saved_files",
	"repairs",
	"maintenances",
	"appliances",
	"todos",
	"todo_task_migrations",
}

// dropAllTables drops every application table using raw SQL.
// Works on both SQLite and PostgreSQL — DROP TABLE IF EXISTS is ANSI SQL.
func dropAllTables(db *gorm.DB) error {
	// note: CASCADE is Postgres-only; SQLite doesn't support it (and has no FK enforcement by default)
	cascade := ""
	if db.Dialector.Name() == dialectPostgres {
		cascade = " CASCADE"
	}
	for _, table := range tableDropOrder {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s%s", table, cascade)).Error; err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}
	return nil
}

// tablesWithSequences are the subset of tableDropOrder that have a SERIAL/BIGSERIAL id column.
// todo_task_migrations uses BIGINT PRIMARY KEY (no sequence), so it's excluded.
// note: Postgres only — sequences don't exist in SQLite.
var tablesWithSequences = []string{
	"tasks",
	"notes",
	"saved_files",
	"repairs",
	"maintenances",
	"appliances",
	"todos",
}

func resetPostgresSequences(db *gorm.DB) error {
	if db.Dialector.Name() != dialectPostgres {
		return nil
	}
	for _, table := range tablesWithSequences {
		sql := fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE(MAX(id), 0) + 1, false) FROM %s`,
			table, table,
		)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("reset sequence %s: %w", table, err)
		}
	}
	return nil
}

// ImportFromJSON replaces all DB data with the payload contents.
// Steps: drop all tables → re-migrate → bulk insert from payload.
// uploadsDir is the directory containing extracted upload files (may be "").
func ImportFromJSON(db *gorm.DB, payload *models.BackupPayload, uploadsDir string) (*models.ImportResult, error) {
	result := &models.ImportResult{}

	// 1. Drop all tables
	if err := dropAllTables(db); err != nil {
		return nil, fmt.Errorf("drop tables: %w", err)
	}

	// 2. Re-create schema
	if err := MigrateGorm(db); err != nil {
		return nil, fmt.Errorf("re-migrate: %w", err)
	}

	// 3. Insert in FK dependency order (parents before children)
	// note: insert individually (not batch) so GORM handles mixed auto/explicit IDs correctly.
	insertEach := func(name string, fn func(i int) error, count int) {
		for i := 0; i < count; i++ {
			if err := fn(i); err != nil {
				result.Errors++
				result.ErrorMessage += fmt.Sprintf("%s[%d]: %v; ", name, i, err)
			} else {
				result.Inserted++
			}
		}
	}

	insertEach("Appliance", func(i int) error { return db.Create(&payload.Entities.Appliances[i]).Error }, len(payload.Entities.Appliances))
	insertEach("Todo", func(i int) error { return db.Create(&payload.Entities.Todos[i]).Error }, len(payload.Entities.Todos))
	insertEach("Maintenance", func(i int) error { return db.Create(&payload.Entities.Maintenance[i]).Error }, len(payload.Entities.Maintenance))
	insertEach("Repair", func(i int) error { return db.Create(&payload.Entities.Repairs[i]).Error }, len(payload.Entities.Repairs))
	insertEach("SavedFile", func(i int) error { return db.Create(&payload.Entities.SavedFiles[i]).Error }, len(payload.Entities.SavedFiles))
	insertEach("Note", func(i int) error { return db.Create(&payload.Entities.Notes[i]).Error }, len(payload.Entities.Notes))
	insertEach("Task", func(i int) error { return db.Create(&payload.Entities.Tasks[i]).Error }, len(payload.Entities.Tasks))

	if result.Errors > 0 {
		return result, fmt.Errorf("import errors: %s", result.ErrorMessage)
	}

	// 4. Resync Postgres sequences — inserting explicit IDs doesn't advance them
	if err := resetPostgresSequences(db); err != nil {
		return result, fmt.Errorf("reset sequences: %w", err)
	}

	// 5. If both todos and tasks were imported, mark all todos as already migrated
	// so MigrateTodosToTasks on next startup doesn't create duplicate tasks.
	if len(payload.Entities.Todos) > 0 && len(payload.Entities.Tasks) > 0 {
		for _, todo := range payload.Entities.Todos {
			if err := db.Exec(
				"INSERT OR IGNORE INTO todo_task_migrations (todo_id) VALUES (?)",
				todo.ID,
			).Error; err != nil {
				result.ErrorMessage += fmt.Sprintf("track migration todo[%d]: %v; ", todo.ID, err)
			}
		}
	}

	return result, nil
}

// ImportFromJSONFile reads a JSON file and delegates to ImportFromJSON.
func ImportFromJSONFile(db *gorm.DB, jsonFilePath string, uploadsDir string) (*models.ImportResult, error) {
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", jsonFilePath, err)
	}

	var payload models.BackupPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal backup: %w", err)
	}

	return ImportFromJSON(db, &payload, uploadsDir)
}

// ImportUploads wipes the uploads directory and copies files from extractedUploadsPath.
func ImportUploads(extractedUploadsPath string) error {
	appUploadsRoot := "./data/uploads"

	if err := os.RemoveAll(appUploadsRoot); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove uploads: %w", err)
	}
	if err := os.MkdirAll(appUploadsRoot, 0755); err != nil {
		return fmt.Errorf("mkdir uploads: %w", err)
	}

	if extractedUploadsPath == "" {
		return nil // no uploads in backup, that's fine
	}

	return filepath.Walk(extractedUploadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, err := filepath.Rel(extractedUploadsPath, path)
		if err != nil {
			return err
		}

		dest := filepath.Join(appUploadsRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		return err
	})
}
