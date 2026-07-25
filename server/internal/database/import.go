package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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

// importLogDDL creates the import_log tracking table.
// Not in tableDropOrder — never dropped, persists across imports.
// Used to detect interrupted imports on server restart.
const importLogDDL = `CREATE TABLE IF NOT EXISTS import_log (
    id           TEXT PRIMARY KEY,
    status       TEXT NOT NULL DEFAULT 'in_progress',
    error_msg    TEXT,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
)`

func ensureImportLogTable(db *gorm.DB) error {
	return db.Exec(importLogDDL).Error
}

// CompleteImport marks an import as completed after upload swap succeeds.
// Best-effort — errors are logged but not returned (uploads already swapped).
func CompleteImport(db *gorm.DB, importID string) {
	if err := db.Exec("UPDATE import_log SET status = 'completed', completed_at = CURRENT_TIMESTAMP WHERE id = ?", importID).Error; err != nil {
		fmt.Printf("Warning: failed to mark import %s as completed: %v\n", importID, err)
	}
}

// FailImport marks an import as failed when upload swap errors out.
func FailImport(db *gorm.DB, importID, reason string) {
	if err := db.Exec("UPDATE import_log SET status = 'failed', error_msg = ? WHERE id = ?", reason, importID).Error; err != nil {
		fmt.Printf("Warning: failed to mark import %s as failed: %v\n", importID, err)
	}
}

// CheckImportLog checks for incomplete imports and returns a warning message
// if any are found. Should be called at server startup after MigrateGorm.
func CheckImportLog(db *gorm.DB) string {
	if err := ensureImportLogTable(db); err != nil {
		return fmt.Sprintf("Warning: could not ensure import_log table: %v", err)
	}
	type pendingImport struct {
		ID        string
		Status    string
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var pending []pendingImport
	if err := db.Table("import_log").Where("status NOT IN ('completed', 'failed')").Find(&pending).Error; err != nil {
		return fmt.Sprintf("Warning: could not check import log: %v", err)
	}
	if len(pending) == 0 {
		return ""
	}
	var msg string
	for _, p := range pending {
		msg += fmt.Sprintf("Import %s was interrupted (status: %s, started: %s). Re-run import to ensure data consistency.\n", p.ID, p.Status, p.CreatedAt.Format(time.RFC3339))
	}
	return msg
}

// validatePayload checks basic structural and required-field constraints on the
// backup payload, before any destructive DB operations begin.
// Returns the first error found — fail-fast.
func validatePayload(payload *models.BackupPayload) error {
	if payload.Version == "" {
		return fmt.Errorf("backup version is required")
	}
	if payload.DatabaseType == "" {
		return fmt.Errorf("database type is required")
	}

	seenIDs := make(map[uint]bool)

	for i, e := range payload.Entities.Appliances {
		if e.ApplianceName == "" {
			return fmt.Errorf("appliance[%d].applianceName: must not be empty", i)
		}
		if e.Manufacturer == "" {
			return fmt.Errorf("appliance[%d].manufacturer: must not be empty", i)
		}
		if e.ModelNumber == "" {
			return fmt.Errorf("appliance[%d].modelNumber: must not be empty", i)
		}
		if e.SerialNumber == "" {
			return fmt.Errorf("appliance[%d].serialNumber: must not be empty", i)
		}
		if e.YearPurchased == "" {
			return fmt.Errorf("appliance[%d].yearPurchased: must not be empty", i)
		}
		if e.PurchasePrice == "" {
			return fmt.Errorf("appliance[%d].purchasePrice: must not be empty", i)
		}
		if e.Location == "" {
			return fmt.Errorf("appliance[%d].location: must not be empty", i)
		}
		if e.Type == "" {
			return fmt.Errorf("appliance[%d].type: must not be empty", i)
		}
		if e.ID != 0 {
			if seenIDs[e.ID] {
				return fmt.Errorf("duplicate appliance ID: %d", e.ID)
			}
			seenIDs[e.ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i, e := range payload.Entities.Todos {
		if e.UserID == "" {
			return fmt.Errorf("todo[%d].userid: must not be empty", i)
		}
		if e.ID != 0 {
			if seenIDs[e.ID] {
				return fmt.Errorf("duplicate todo ID: %d", e.ID)
			}
			seenIDs[e.ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i, e := range payload.Entities.Maintenance {
		if e.Description == "" {
			return fmt.Errorf("maintenance[%d].description: must not be empty", i)
		}
		if e.Date == "" {
			return fmt.Errorf("maintenance[%d].date: must not be empty", i)
		}
		if e.ID != 0 {
			if seenIDs[e.ID] {
				return fmt.Errorf("duplicate maintenance ID: %d", e.ID)
			}
			seenIDs[e.ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i, e := range payload.Entities.Repairs {
		if e.Description == "" {
			return fmt.Errorf("repair[%d].description: must not be empty", i)
		}
		if e.Date == "" {
			return fmt.Errorf("repair[%d].date: must not be empty", i)
		}
		if e.ID != 0 {
			if seenIDs[e.ID] {
				return fmt.Errorf("duplicate repair ID: %d", e.ID)
			}
			seenIDs[e.ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i, e := range payload.Entities.SavedFiles {
		if e.Path == "" {
			return fmt.Errorf("savedFile[%d].path: must not be empty", i)
		}
		if e.OriginalName == "" {
			return fmt.Errorf("savedFile[%d].originalName: must not be empty", i)
		}
		if e.Type == "" {
			return fmt.Errorf("savedFile[%d].type: must not be empty", i)
		}
		if e.UserID == "" {
			return fmt.Errorf("savedFile[%d].userid: must not be empty", i)
		}
		if e.ID != 0 {
			if seenIDs[e.ID] {
				return fmt.Errorf("duplicate savedFile ID: %d", e.ID)
			}
			seenIDs[e.ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i := range payload.Entities.Notes {
		if payload.Entities.Notes[i].ID != 0 {
			if seenIDs[payload.Entities.Notes[i].ID] {
				return fmt.Errorf("duplicate note ID: %d", payload.Entities.Notes[i].ID)
			}
			seenIDs[payload.Entities.Notes[i].ID] = true
		}
	}

	seenIDs = make(map[uint]bool)
	for i := range payload.Entities.Tasks {
		if payload.Entities.Tasks[i].ID != 0 {
			if seenIDs[payload.Entities.Tasks[i].ID] {
				return fmt.Errorf("duplicate task ID: %d", payload.Entities.Tasks[i].ID)
			}
			seenIDs[payload.Entities.Tasks[i].ID] = true
		}
	}

	return nil
}

// ImportFromJSON replaces all DB data with the payload contents.
// Steps: drop all tables → re-migrate → bulk insert from payload.
// The critical path (drop → migrate → insert → reset sequences) is wrapped
// in a database transaction so that any failure fully rolls back, preventing
// a broken or empty database state.
// uploadsDir is the directory containing extracted upload files (may be "").
func ImportFromJSON(db *gorm.DB, payload *models.BackupPayload, uploadsDir string) (*models.ImportResult, error) {
	result := &models.ImportResult{}

	if err := ensureImportLogTable(db); err != nil {
		return nil, fmt.Errorf("ensure import_log: %w", err)
	}

	sanitizeFKs(payload)

	if err := validatePayload(payload); err != nil {
		return nil, fmt.Errorf("invalid backup: %w", err)
	}

	importID := fmt.Sprintf("imp_%d", time.Now().UnixNano())
	result.ImportID = importID

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Drop all tables
		if err := dropAllTables(tx); err != nil {
			return fmt.Errorf("drop tables: %w", err)
		}

		// 2. Re-create schema
		if err := MigrateGorm(tx); err != nil {
			return fmt.Errorf("re-migrate: %w", err)
		}

		// 3. Insert in FK dependency order (parents before children)
		// note: insert individually (not batch) so GORM handles mixed auto/explicit IDs correctly.
		insertEach := func(name string, fn func(i int) error, count int) error {
			for i := 0; i < count; i++ {
				if err := fn(i); err != nil {
					return fmt.Errorf("%s[%d]: %w", name, i, err)
				}
				result.Inserted++
			}
			return nil
		}

		if err := insertEach("Appliance", func(i int) error { return tx.Create(&payload.Entities.Appliances[i]).Error }, len(payload.Entities.Appliances)); err != nil {
			return err
		}
		if err := insertEach("Todo", func(i int) error { return tx.Create(&payload.Entities.Todos[i]).Error }, len(payload.Entities.Todos)); err != nil {
			return err
		}
		if err := insertEach("Maintenance", func(i int) error { return tx.Create(&payload.Entities.Maintenance[i]).Error }, len(payload.Entities.Maintenance)); err != nil {
			return err
		}
		if err := insertEach("Repair", func(i int) error { return tx.Create(&payload.Entities.Repairs[i]).Error }, len(payload.Entities.Repairs)); err != nil {
			return err
		}
		if err := insertEach("SavedFile", func(i int) error { return tx.Create(&payload.Entities.SavedFiles[i]).Error }, len(payload.Entities.SavedFiles)); err != nil {
			return err
		}
		if err := insertEach("Note", func(i int) error { return tx.Create(&payload.Entities.Notes[i]).Error }, len(payload.Entities.Notes)); err != nil {
			return err
		}
		if err := insertEach("Task", func(i int) error { return tx.Create(&payload.Entities.Tasks[i]).Error }, len(payload.Entities.Tasks)); err != nil {
			return err
		}

		// 4. Resync Postgres sequences — inserting explicit IDs doesn't advance them
		if err := resetPostgresSequences(tx); err != nil {
			return fmt.Errorf("reset sequences: %w", err)
		}

		// 5. Record in_progress in import_log for crash detection.
		// Clear old entries first — only care about latest.
		if err := tx.Exec("DELETE FROM import_log").Error; err != nil {
			return fmt.Errorf("clear import_log: %w", err)
		}
		if err := tx.Exec("INSERT INTO import_log (id, status) VALUES (?, 'in_progress')", importID).Error; err != nil {
			return fmt.Errorf("record import state: %w", err)
		}

		// 6. Ensure todo→task migration tracking table and migrate.
		// Runs inside the transaction so that any failure rolls back the entire import.
		if len(payload.Entities.Todos) > 0 {
			if err := tx.Exec(`CREATE TABLE IF NOT EXISTS todo_task_migrations (todo_id BIGINT PRIMARY KEY)`).Error; err != nil {
				return fmt.Errorf("create migration tracking table: %w", err)
			}
			// If both todos and tasks were imported, mark all todos as already
			// migrated so MigrateTodosToTasks doesn't create duplicates.
			if len(payload.Entities.Tasks) > 0 {
				insertSQL := "INSERT OR IGNORE INTO todo_task_migrations (todo_id) VALUES (?)"
				if tx.Dialector.Name() == dialectPostgres {
					insertSQL = "INSERT INTO todo_task_migrations (todo_id) VALUES (?) ON CONFLICT (todo_id) DO NOTHING"
				}
				for _, todo := range payload.Entities.Todos {
					if err := tx.Exec(insertSQL, todo.ID).Error; err != nil {
						return fmt.Errorf("track migration todo[%d]: %w", todo.ID, err)
					}
				}
			}
			// Migrate any remaining todos (those without a tracking entry) to tasks now.
			if err := MigrateTodosToTasks(tx); err != nil {
				return fmt.Errorf("migrate todos: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
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

// ImportUploads replaces the uploads directory with files from extractedUploadsPath.
// Files are staged in a temp directory first, then atomically swapped into place
// so that a partial copy failure does not wipe the existing uploads.
func ImportUploads(extractedUploadsPath string) error {
	appUploadsRoot := "./data/uploads"

	if extractedUploadsPath == "" {
		return nil
	}

	tempDir, err := os.MkdirTemp("./data", "uploads-import-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	}()

	err = filepath.Walk(extractedUploadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, err := filepath.Rel(extractedUploadsPath, path)
		if err != nil {
			return err
		}

		dest := filepath.Join(tempDir, rel)
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
	if err != nil {
		return fmt.Errorf("stage uploads: %w", err)
	}

	// Backup old uploads by renaming out of the way.
	// This keeps the original data intact in case the swap fails.
	oldBackup := appUploadsRoot + ".bak"
	if err := os.RemoveAll(oldBackup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean stale backup: %w", err)
	}

	if err := os.Rename(appUploadsRoot, oldBackup); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("backup old uploads: %w", err)
		}
		// No existing uploads directory — that's fine.
		oldBackup = ""
	}

	if err := os.Rename(tempDir, appUploadsRoot); err != nil {
		// Swap failed — restore original uploads if we backed them up.
		if oldBackup != "" {
			_ = os.Rename(oldBackup, appUploadsRoot)
		}
		return fmt.Errorf("swap uploads: %w", err)
	}

	// Success — clean up old backup and mark temp as handled.
	if oldBackup != "" {
		_ = os.RemoveAll(oldBackup)
	}
	tempDir = ""
	return nil
}
