package main

import (
	"encoding/json"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestImportLockMiddleware(t *testing.T) {
	t.Run("not importing passes through", func(t *testing.T) {
		var importing atomic.Bool
		app := fiber.New()
		app.Use(ImportLockMiddleware(&importing))

		app.Get("/api/appliances", func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/api/appliances", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("importing blocks non-critical routes", func(t *testing.T) {
		var importing atomic.Bool
		importing.Store(true)
		app := fiber.New()
		app.Use(ImportLockMiddleware(&importing))

		app.Get("/api/appliances", func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/api/appliances", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		if body["status"] != "busy" {
			t.Errorf("expected status 'busy', got %v", body["status"])
		}
	})

	t.Run("health bypasses import lock", func(t *testing.T) {
		var importing atomic.Bool
		importing.Store(true)
		app := fiber.New()
		app.Use(ImportLockMiddleware(&importing))

		app.Get("/api/health", func(c fiber.Ctx) error {
			return c.SendString("healthy")
		})

		req := httptest.NewRequest("GET", "/api/health", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("import endpoint bypasses lock", func(t *testing.T) {
		var importing atomic.Bool
		importing.Store(true)
		app := fiber.New()
		app.Use(ImportLockMiddleware(&importing))

		app.Post("/api/backup/import", func(c fiber.Ctx) error {
			return c.SendString("importing")
		})

		req := httptest.NewRequest("POST", "/api/backup/import", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode == fiber.StatusServiceUnavailable {
			t.Error("import endpoint should bypass the lock")
		}
	})

	t.Run("non-api routes bypass lock", func(t *testing.T) {
		var importing atomic.Bool
		importing.Store(true)
		app := fiber.New()
		app.Use(ImportLockMiddleware(&importing))

		app.Get("/", func(c fiber.Ctx) error {
			return c.SendString("home")
		})

		req := httptest.NewRequest("GET", "/", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}


