package database

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

func ensureImportLogTableForTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := ensureImportLogTable(db); err != nil {
		t.Fatalf("failed to create import_log table: %v", err)
	}
}

func TestCompleteImport(t *testing.T) {
	db := TestDB(t)

	ensureImportLogTableForTest(t, db)

	importID := "test-complete-001"
	if err := db.Exec("INSERT INTO import_log (id, status) VALUES (?, 'in_progress')", importID).Error; err != nil {
		t.Fatalf("failed to insert test row: %v", err)
	}

	CompleteImport(db, importID)

	type importLogRow struct {
		ID     string
		Status string
	}
	var row importLogRow
	if err := db.Table("import_log").Where("id = ?", importID).First(&row).Error; err != nil {
		t.Fatalf("failed to query import_log: %v", err)
	}

	if row.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", row.Status)
	}

	var count int64
	db.Table("import_log").Where("id = ? AND completed_at IS NOT NULL", importID).Count(&count)
	if count == 0 {
		t.Error("expected completed_at to be set, but it IS NULL")
	}
}

func TestFailImport(t *testing.T) {
	db := TestDB(t)

	ensureImportLogTableForTest(t, db)

	importID := "test-fail-001"
	if err := db.Exec("INSERT INTO import_log (id, status) VALUES (?, 'in_progress')", importID).Error; err != nil {
		t.Fatalf("failed to insert test row: %v", err)
	}

	reason := "deliberate test failure"
	FailImport(db, importID, reason)

	type importLogRow struct {
		ID       string
		Status   string
		ErrorMsg *string `gorm:"column:error_msg"`
	}
	var row importLogRow
	if err := db.Table("import_log").Where("id = ?", importID).First(&row).Error; err != nil {
		t.Fatalf("failed to query import_log: %v", err)
	}

	if row.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", row.Status)
	}
	if row.ErrorMsg == nil {
		t.Fatal("expected error_msg to be set, got nil")
	}
	if *row.ErrorMsg != reason {
		t.Errorf("expected error_msg %q, got %q", reason, *row.ErrorMsg)
	}
}

func TestCheckImportLog(t *testing.T) {
	t.Run("empty log returns empty string", func(t *testing.T) {
		db := TestDB(t)
		msg := CheckImportLog(db)
		if msg != "" {
			t.Errorf("expected empty string, got %q", msg)
		}
	})

	t.Run("all completed or failed returns empty", func(t *testing.T) {
		db := TestDB(t)
		ensureImportLogTableForTest(t, db)

		if err := db.Exec("INSERT INTO import_log (id, status) VALUES ('c1', 'completed')").Error; err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
		if err := db.Exec("INSERT INTO import_log (id, status) VALUES ('f1', 'failed')").Error; err != nil {
			t.Fatalf("failed to insert: %v", err)
		}

		msg := CheckImportLog(db)
		if msg != "" {
			t.Errorf("expected empty string, got %q", msg)
		}
	})

	t.Run("stale in_progress returns warning", func(t *testing.T) {
		db := TestDB(t)
		ensureImportLogTableForTest(t, db)

		if err := db.Exec("INSERT INTO import_log (id, status) VALUES ('stale-001', 'in_progress')").Error; err != nil {
			t.Fatalf("failed to insert: %v", err)
		}

		msg := CheckImportLog(db)
		if msg == "" {
			t.Fatal("expected non-empty warning for in_progress import")
		}
		if msg != "" && !strings.Contains(msg, "stale-001") {
			t.Errorf("expected warning to mention 'stale-001', got %q", msg)
		}
	})
}
