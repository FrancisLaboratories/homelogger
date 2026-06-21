package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withEnv(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if value == "" {
		_ = os.Unsetenv(key)
	} else {
		_ = os.Setenv(key, value)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestConnectGorm_LocksDialectOnFirstSuccessfulRun(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, ".db_dialect")
	dbPath := filepath.Join(tmp, "test.db")

	withEnv(t, "DB_DIALECT", "sqlite")
	withEnv(t, "DB_DIALECT_LOCK_PATH", lockPath)
	withEnv(t, "DEMO_DB_PATH", dbPath)
	withEnv(t, "DATABASE_URL", "")
	withEnv(t, "FORCE_DB_DIALECT_CHANGE", "")

	db, err := ConnectGorm()
	if err != nil {
		t.Fatalf("first ConnectGorm failed: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("expected lock file: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != dialectSQLite {
		t.Fatalf("expected locked dialect %q, got %q", dialectSQLite, got)
	}

	withEnv(t, "DB_DIALECT", "postgres")
	if _, err := ConnectGorm(); err == nil {
		t.Fatalf("expected lock mismatch error")
	} else if !strings.Contains(err.Error(), "cannot be changed after first run") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectDialect_ForceOverrideAllowsMismatch(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, ".db_dialect")
	if err := os.WriteFile(lockPath, []byte(dialectSQLite), 0644); err != nil {
		t.Fatalf("write lock file: %v", err)
	}

	withEnv(t, "DB_DIALECT", "postgres")
	withEnv(t, "DB_DIALECT_LOCK_PATH", lockPath)
	withEnv(t, "FORCE_DB_DIALECT_CHANGE", "true")

	selection, err := selectDialect()
	if err != nil {
		t.Fatalf("selectDialect failed: %v", err)
	}
	if selection.dialect != dialectPostgres {
		t.Fatalf("expected selected %q, got %q", dialectPostgres, selection.dialect)
	}
	if !selection.lockExists || selection.lockedDialect != dialectSQLite {
		t.Fatalf("unexpected lock metadata: %+v", selection)
	}
}
