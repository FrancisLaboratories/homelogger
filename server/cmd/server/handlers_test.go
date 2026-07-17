package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// openTestDB delegates to database.TestDB — uses SQLite by default, Postgres when TEST_DB_DIALECT=postgres.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return database.TestDB(t)
}

// createApp registers a minimal set of handlers using the provided DB.
func createApp(db *gorm.DB) *fiber.App {
	app := fiber.New()

	app.Get("/api/health", func(c fiber.Ctx) error {
		sqlDB, err := db.DB()
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error: " + err.Error()
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"db": dbStatus})
	})

	app.Get("/api/appliances", func(c fiber.Ctx) error {
		apps, err := database.GetAppliances(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(apps)
	})

	app.Post("/api/appliances/add", func(c fiber.Ctx) error {
		var body models.Appliance
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		a, err := database.AddAppliance(db, &body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(a)
	})

	app.Post("/api/todo/add", func(c fiber.Ctx) error {
		var body struct {
			Label   string `json:"label"`
			Checked bool   `json:"checked"`
			UserID  string `json:"userid"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		tdo, err := database.AddTodo(db, body.Label, body.Checked, body.UserID, 0, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(tdo)
	})

	app.Get("/api/todo", func(c fiber.Ctx) error {
		todos, err := database.GetTodos(db, 0, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(todos)
	})

	return app
}

func TestApplianceEndpoints(t *testing.T) {
	db := openTestDB(t)
	app := createApp(db)

	// initially empty
	req := httptest.NewRequest("GET", "/api/appliances", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// add appliance
	body := models.Appliance{ApplianceName: "Xfridge", Manufacturer: "Acme"}
	b, _ := json.Marshal(body)
	req = httptest.NewRequest("POST", "/api/appliances/add", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 on add, got %d", resp.StatusCode)
	}

	// list should now have one
	req = httptest.NewRequest("GET", "/api/appliances", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestTodoEndpoints(t *testing.T) {
	db := openTestDB(t)
	app := createApp(db)

	// add todo
	payload := map[string]interface{}{"label": "t1", "checked": false, "userid": "1"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/todo/add", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 on add todo, got %d", resp.StatusCode)
	}

	// get todos
	req = httptest.NewRequest("GET", "/api/todo", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 on get todos, got %d", resp.StatusCode)
	}
}

func createAppWithMaintenanceRepair(db *gorm.DB) *fiber.App {
	app := fiber.New()

	app.Post("/api/maintenance/add", maintenanceAddHandler(db))
	app.Put("/api/maintenance/update/:id", maintenanceUpdateHandler(db))
	app.Delete("/api/maintenance/delete/:id", maintenanceDeleteHandler(db))

	app.Post("/api/repair/add", repairAddHandler(db))
	app.Put("/api/repair/update/:id", repairUpdateHandler(db))
	app.Delete("/api/repair/delete/:id", repairDeleteHandler(db))

	return app
}

// parseIDParam parses a numeric :id param and returns it as uint
func parseIDParam(c fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func maintenanceAddHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		var body struct {
			models.Maintenance
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		m, err := database.AddMaintenance(db, &body.Maintenance)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.Status(fiber.StatusCreated).JSON(m)
	}
}

func maintenanceUpdateHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}
		var body struct {
			Description string  `json:"description"`
			Date        string  `json:"date"`
			Cost        float64 `json:"cost"`
			Notes       string  `json:"notes"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		updated, err := database.UpdateMaintenance(db, id, body.Description, body.Date, body.Cost, body.Notes)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(updated)
	}
}

func maintenanceDeleteHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}
		if err := database.DeleteMaintenance(db, id); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func repairAddHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		var body struct {
			models.Repair
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		r, err := database.AddRepair(db, &body.Repair)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.Status(fiber.StatusCreated).JSON(r)
	}
}

func repairUpdateHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}
		var body struct {
			Description string  `json:"description"`
			Date        string  `json:"date"`
			Cost        float64 `json:"cost"`
			Notes       string  `json:"notes"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad")
		}
		updated, err := database.UpdateRepair(db, id, body.Description, body.Date, body.Cost, body.Notes)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.JSON(updated)
	}
}

func repairDeleteHandler(db *gorm.DB) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}
		if err := database.DeleteRepair(db, id); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("err")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func TestMaintenanceEndpoints(t *testing.T) {
	db := openTestDB(t)
	app := createAppWithMaintenanceRepair(db)

	// Add a maintenance record
	payload := map[string]interface{}{
		"description":   "Replace filter",
		"date":          "2026-03-01",
		"cost":          50.0,
		"notes":         "Annual filter change",
		"referenceType": "Space",
		"spaceType":     "HVAC",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/maintenance/add", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 on add, got %d", resp.StatusCode)
	}
	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	id := int(created["id"].(float64))

	// Update the maintenance record
	update := map[string]interface{}{
		"description": "Replace filter - updated",
		"date":        "2026-04-01",
		"cost":        75.0,
		"notes":       "Updated notes",
	}
	b, _ = json.Marshal(update)
	req = httptest.NewRequest("PUT", "/api/maintenance/update/"+strconv.Itoa(id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}
	var updated map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated["description"] != "Replace filter - updated" {
		t.Errorf("expected updated description, got %v", updated["description"])
	}
	if updated["cost"] != 75.0 {
		t.Errorf("expected updated cost 75.0, got %v", updated["cost"])
	}

	// Delete the maintenance record
	req = httptest.NewRequest("DELETE", "/api/maintenance/delete/"+strconv.Itoa(id), nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204 on delete, got %d", resp.StatusCode)
	}

	// Update a non-existent record should fail
	b, _ = json.Marshal(update)
	req = httptest.NewRequest("PUT", "/api/maintenance/update/"+strconv.Itoa(id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500 on update of deleted record, got %d", resp.StatusCode)
	}
}

func TestRepairEndpoints(t *testing.T) {
	db := openTestDB(t)
	app := createAppWithMaintenanceRepair(db)

	// Add a repair record
	payload := map[string]interface{}{
		"description":   "Fix leak",
		"date":          "2026-03-15",
		"cost":          120.0,
		"notes":         "Kitchen sink",
		"referenceType": "Space",
		"spaceType":     "Plumbing",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/repair/add", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 on add, got %d", resp.StatusCode)
	}
	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	id := int(created["id"].(float64))

	// Update the repair record
	update := map[string]interface{}{
		"description": "Fix leak - updated",
		"date":        "2026-04-15",
		"cost":        200.0,
		"notes":       "Updated repair notes",
	}
	b, _ = json.Marshal(update)
	req = httptest.NewRequest("PUT", "/api/repair/update/"+strconv.Itoa(id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}
	var updated map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated["description"] != "Fix leak - updated" {
		t.Errorf("expected updated description, got %v", updated["description"])
	}
	if updated["cost"] != 200.0 {
		t.Errorf("expected updated cost 200.0, got %v", updated["cost"])
	}

	// Delete the repair record
	req = httptest.NewRequest("DELETE", "/api/repair/delete/"+strconv.Itoa(id), nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204 on delete, got %d", resp.StatusCode)
	}

	// Update a non-existent record should fail
	b, _ = json.Marshal(update)
	req = httptest.NewRequest("PUT", "/api/repair/update/"+strconv.Itoa(id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500 on update of deleted record, got %d", resp.StatusCode)
	}
}
