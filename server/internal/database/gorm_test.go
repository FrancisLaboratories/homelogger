package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDatabaseConfig_DefaultsToSQLite(t *testing.T) {
	tmpDir, err := os.MkdirTemp(".", "dbenginemeta")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Clearenv()
	os.Setenv(envEngineMetadataPath, filepath.Join(tmpDir, defaultEngineTXTName))

	config, err := parseDatabaseConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.Driver != "sqlite" {
		t.Fatalf("expected sqlite driver, got %s", config.Driver)
	}
	if config.DSN != defaultSQLiteDBPath {
		t.Fatalf("expected default sqlite path %s, got %s", defaultSQLiteDBPath, config.DSN)
	}
}

func TestParseDatabaseConfig_UsesPostgresDSN(t *testing.T) {
	tmpDir, err := os.MkdirTemp(".", "dbenginemeta")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Clearenv()
	os.Setenv(envEngineMetadataPath, filepath.Join(tmpDir, defaultEngineTXTName))
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/homelogger?sslmode=disable")

	config, err := parseDatabaseConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.Driver != "postgres" {
		t.Fatalf("expected postgres driver, got %s", config.Driver)
	}
}

func TestParseDatabaseConfig_RequiresDSNForPostgres(t *testing.T) {
	tmpDir, err := os.MkdirTemp(".", "dbenginemeta")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Clearenv()
	os.Setenv(envEngineMetadataPath, filepath.Join(tmpDir, defaultEngineTXTName))
	os.Setenv("DB_TYPE", "postgres")

	_, err = parseDatabaseConfig()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing for postgres")
	}
}

func TestParseDatabaseConfig_UsesMariaDBAlias(t *testing.T) {
	tmpDir, err := os.MkdirTemp(".", "dbenginemeta")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Clearenv()
	os.Setenv(envEngineMetadataPath, filepath.Join(tmpDir, defaultEngineTXTName))
	os.Setenv("DB_TYPE", "mysql")
	os.Setenv("DATABASE_URL", "user:pass@tcp(localhost:3306)/homelogger")

	config, err := parseDatabaseConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.Driver != "mariadb" {
		t.Fatalf("expected mariadb driver, got %s", config.Driver)
	}
}

func TestEngineMetadataIsPinned(t *testing.T) {
	tmpDir := t.TempDir()
	engineFile := filepath.Join(tmpDir, defaultEngineTXTName)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	if err := os.WriteFile(engineFile, []byte("sqlite"), 0644); err != nil {
		t.Fatalf("failed to write engine file: %v", err)
	}

	if err := ensureEngineMetadata("postgres", engineFile); err == nil {
		t.Fatal("expected error when engine metadata conflicts")
	}
}
