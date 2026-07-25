package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

type testAppConfig struct {
	db        *gorm.DB
	importing *atomic.Bool
	backupMu  *sync.Mutex
}

func createTestApp(cfg testAppConfig) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 100 * 1024 * 1024})

	app.Use(ImportLockMiddleware(cfg.importing))

	api := app.Group("/api")

	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "importing": cfg.importing.Load()})
	})

	api.Post("/backup/import", ImportBackupHandler(cfg.db, cfg.importing, cfg.backupMu))

	api.Get("/appliances", func(c fiber.Ctx) error {
		var apps []models.Appliance
		if err := cfg.db.Find(&apps).Error; err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(apps)
	})

	return app
}

func createTestBackupZIP(t *testing.T, payload *models.BackupPayload) ([]byte, string) {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := w.Create("data.json")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write data.json: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	return buf.Bytes(), "test-backup.zip"
}

func multipartRequest(url string, zipData []byte, filename string) *http.Request {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, _ := w.CreateFormFile("backup", filename)
	fw.Write(zipData)
	w.Close()

	req := httptest.NewRequest("POST", url, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestImportHandler_Success(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "Test Fridge"},
			},
		},
	}

	zipData, filename := createTestBackupZIP(t, payload)
	req := multipartRequest("/api/backup/import", zipData, filename)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", body["status"])
	}
	if body["importId"] == nil || body["importId"] == "" {
		t.Error("expected non-empty importId")
	}
	if body["inserted"] != float64(1) {
		t.Errorf("expected inserted 1, got %v", body["inserted"])
	}

	// Verify data was actually imported
	var count int64
	db.Model(&models.Appliance{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 appliance in DB, got %d", count)
	}
}

func TestImportHandler_NoFile(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	req := httptest.NewRequest("POST", "/api/backup/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestImportHandler_NestedDataJSON(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "Fridge"},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("subdir/data.json")
	f.Write(data)
	w.Close()

	req := multipartRequest("/api/backup/import", buf.Bytes(), "nested.zip")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for nested data.json, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "failed" {
		t.Errorf("expected status 'failed', got %v", body["status"])
	}
}

func TestImportHandler_MissingDataJSON(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("some-file.txt")
	f.Write([]byte("hello"))
	w.Close()

	req := multipartRequest("/api/backup/import", buf.Bytes(), "no-data.zip")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for missing data.json, got %d", resp.StatusCode)
	}
}

func TestImportHandler_MalformedZIP(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	// Send garbage instead of a valid ZIP
	req := multipartRequest("/api/backup/import", []byte("not-a-zip-file"), "bad.zip")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 for malformed zip, got %d", resp.StatusCode)
	}
}

func TestImportHandler_WithUploads(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool
	var mu sync.Mutex

	app := createTestApp(testAppConfig{
		db:        db,
		importing: &importing,
		backupMu:  &mu,
	})

	payload := &models.BackupPayload{
		Version:      database.BackupVersion,
		DatabaseType: db.Dialector.Name(),
		Entities: models.Entities{
			Appliances: []models.Appliance{
				{ApplianceName: "Fridge"},
			},
			SavedFiles: []models.SavedFile{
				{Path: "./data/uploads/photo.jpg", OriginalName: "photo.jpg", UserID: "u1"},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, _ := w.Create("data.json")
	f.Write(data)

	f, _ = w.Create("uploads/photo.jpg")
	f.Write([]byte("photo-data"))

	w.Close()

	// Need ./data to exist for ImportUploads
	if err := os.MkdirAll("./data", 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}

	req := multipartRequest("/api/backup/import", buf.Bytes(), "with-uploads.zip")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d. Response: %s", resp.StatusCode, readBody(resp))
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "completed" {
		t.Errorf("expected 'completed', got %v", body["status"])
	}

	// Clean up the uploaded files
	os.RemoveAll("./data/uploads")
	os.RemoveAll("./data/uploads.bak")

	// Also clean up any temp directories created by ImportUploads
	tempDirs, _ := filepath.Glob("./data/uploads-import-*")
	for _, d := range tempDirs {
		os.RemoveAll(d)
	}
}

func readBody(r *http.Response) string {
	b, _ := io.ReadAll(r.Body)
	return string(b)
}
