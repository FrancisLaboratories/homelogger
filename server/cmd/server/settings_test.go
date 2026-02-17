package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type settingsResponse struct {
	Locale            string `json:"locale"`
	Language          string `json:"language"`
	Currency          string `json:"currency"`
	TimeZone          string `json:"timeZone"`
	MeasurementSystem string `json:"measurementSystem"`
	WeekStart         int    `json:"weekStart"`
	DateFormat        string `json:"dateFormat"`
	NumberingSystem   string `json:"numberingSystem"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateGorm(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if _, err := database.EnsureSettings(db); err != nil {
		t.Fatalf("ensure settings: %v", err)
	}
	return db
}

func setupTestApp(db *gorm.DB) *fiber.App {
	app := fiber.New()
	registerSettingsRoutes(app, db)
	return app
}

func TestEnsureSettingsCreatesDefaults(t *testing.T) {
	db := setupTestDB(t)

	settings, err := database.GetSettings(db)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	expected := database.DefaultSettings()
	if settings.Locale != expected.Locale ||
		settings.Language != expected.Language ||
		settings.Currency != expected.Currency ||
		settings.TimeZone != expected.TimeZone ||
		settings.MeasurementSystem != expected.MeasurementSystem ||
		settings.WeekStart != expected.WeekStart ||
		settings.DateFormat != expected.DateFormat ||
		settings.NumberingSystem != expected.NumberingSystem {
		t.Fatalf("defaults mismatch: got %+v, expected %+v", settings, expected)
	}
}

func TestSettingsGetEndpoint(t *testing.T) {
	db := setupTestDB(t)
	app := setupTestApp(db)

	req, err := http.NewRequest(http.MethodGet, "/settings", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var body settingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Locale != "en-US" || body.Currency != "USD" || body.TimeZone != "UTC" {
		t.Fatalf("unexpected settings response: %+v", body)
	}
}

func TestSettingsPutEndpoint(t *testing.T) {
	db := setupTestDB(t)
	app := setupTestApp(db)

	payload := map[string]any{
		"locale":            "en-GB",
		"language":          "en",
		"currency":          "GBP",
		"timeZone":          "Europe/London",
		"measurementSystem": "metric",
		"weekStart":         1,
		"dateFormat":        "DD/MM/YYYY",
		"numberingSystem":   "latn",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPut, "/settings", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var updated settingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updated.Currency != "GBP" || updated.TimeZone != "Europe/London" || updated.MeasurementSystem != "metric" || updated.WeekStart != 1 {
		t.Fatalf("unexpected updated settings: %+v", updated)
	}

	settings, err := database.GetSettings(db)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.Currency != "GBP" || settings.TimeZone != "Europe/London" || settings.MeasurementSystem != "metric" || settings.WeekStart != 1 {
		t.Fatalf("db not updated: %+v", settings)
	}
}

func TestSettingsPutValidation(t *testing.T) {
	db := setupTestDB(t)
	app := setupTestApp(db)

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name:    "invalid measurement system",
			payload: map[string]any{"measurementSystem": "unknown"},
		},
		{
			name:    "invalid week start",
			payload: map[string]any{"weekStart": 9},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.payload)
			req, err := http.NewRequest(http.MethodPut, "/settings", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("test request: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("unexpected status: %d", resp.StatusCode)
			}
		})
	}
}

func TestSettingsPutPartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	app := setupTestApp(db)

	initial := models.Settings{
		Locale:            "en-US",
		Language:          "en",
		Currency:          "USD",
		TimeZone:          "UTC",
		MeasurementSystem: "imperial",
		WeekStart:         0,
		DateFormat:        "",
		NumberingSystem:   "latn",
	}
	if err := db.Model(&models.Settings{}).Where("id = ?", 1).Updates(initial).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	payload := map[string]any{
		"currency": "CAD",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPut, "/settings", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var updated settingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updated.Currency != "CAD" || updated.Locale != "en-US" {
		t.Fatalf("partial update unexpected response: %+v", updated)
	}
}
