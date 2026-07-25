package main

import (
	"encoding/json"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
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

	t.Run("real health handler with db", func(t *testing.T) {
		db := openTestDB(t)
		var importing atomic.Bool

		app := fiber.New()
		app.Get("/api/health", HealthHandler(db, false, &importing))

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
	})
}
