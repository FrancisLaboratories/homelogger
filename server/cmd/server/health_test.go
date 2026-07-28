package main

import (
	"encoding/json"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func TestHealthEndpoint_ImportingField(t *testing.T) {
	t.Run("normal state shows importing false", func(t *testing.T) {
		var importing atomic.Bool

		app := fiber.New()
		app.Get("/api/health", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status":    "ok",
				"importing": importing.Load(),
			})
		})

		req := httptest.NewRequest("GET", "/api/health", nil)
		resp, _ := app.Test(req)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		if body["importing"] != false {
			t.Errorf("expected importing=false, got %v", body["importing"])
		}
	})

	t.Run("while importing shows importing true", func(t *testing.T) {
		var importing atomic.Bool
		importing.Store(true)

		app := fiber.New()
		app.Get("/api/health", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status":    "ok",
				"importing": importing.Load(),
			})
		})

		req := httptest.NewRequest("GET", "/api/health", nil)
		resp, _ := app.Test(req)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		if body["importing"] != true {
			t.Errorf("expected importing=true, got %v", body["importing"])
		}
	})
}

func TestHealthEndpoint_RealHandlerWithDB(t *testing.T) {
	db := openTestDB(t)
	var importing atomic.Bool

	app := fiber.New()
	app.Get("/api/health", HealthHandler(func() *gorm.DB { return db }, false, &importing))

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["importing"] != false {
		t.Errorf("expected importing=false, got %v", body["importing"])
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
	if body["db"] != "ok" {
		t.Errorf("expected db ok, got %v", body["db"])
	}
}

func TestHealthEndpoint_UsesLatestDB(t *testing.T) {
	db1 := openTestDB(t)
	db2 := openTestDB(t)
	var importing atomic.Bool
	currentDB := db1

	app := fiber.New()
	app.Get("/api/health", HealthHandler(func() *gorm.DB { return currentDB }, false, &importing))

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 on first request, got %d", resp.StatusCode)
	}

	sqlDB1, err := db1.DB()
	if err != nil {
		t.Fatalf("failed to get sql DB from db1: %v", err)
	}
	_ = sqlDB1.Close()
	currentDB = db2

	req = httptest.NewRequest("GET", "/api/health", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 after swapping DB, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["db"] != "ok" {
		t.Errorf("expected db ok after swap, got %v", body["db"])
	}
}
