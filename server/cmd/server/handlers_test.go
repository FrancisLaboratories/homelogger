package main

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/gofiber/fiber/v2"
    "github.com/masoncfrancis/homelogger/server/internal/database"
    "github.com/masoncfrancis/homelogger/server/internal/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// openTestDB creates an in-memory DB and runs migrations
func openTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file:mem_srv_test?mode=memory&cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed open db: %v", err)
    }
    if err := database.MigrateGorm(db); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    return db
}

// createApp registers a minimal set of handlers using the provided DB.
func createApp(db *gorm.DB) *fiber.App {
    app := fiber.New()

    app.Get("/health", func(c *fiber.Ctx) error {
        sqlDB, err := db.DB()
        dbStatus := "ok"
        if err != nil {
            dbStatus = "error: " + err.Error()
        } else if err := sqlDB.Ping(); err != nil {
            dbStatus = "error: " + err.Error()
        }
        return c.Status(fiber.StatusOK).JSON(fiber.Map{"db": dbStatus})
    })

    app.Get("/appliances", func(c *fiber.Ctx) error {
        apps, err := database.GetAppliances(db)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).SendString("err")
        }
        return c.JSON(apps)
    })

    app.Post("/appliances/add", func(c *fiber.Ctx) error {
        var body models.Appliance
        if err := c.BodyParser(&body); err != nil {
            return c.Status(fiber.StatusBadRequest).SendString("bad")
        }
        a, err := database.AddAppliance(db, &body)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).SendString("err")
        }
        return c.JSON(a)
    })

    app.Post("/todo/add", func(c *fiber.Ctx) error {
        var body struct {
            Label   string `json:"label"`
            Checked bool   `json:"checked"`
            UserID  string `json:"userid"`
        }
        if err := c.BodyParser(&body); err != nil {
            return c.Status(fiber.StatusBadRequest).SendString("bad")
        }
        tdo, err := database.AddTodo(db, body.Label, body.Checked, body.UserID, 0, "")
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).SendString("err")
        }
        return c.JSON(tdo)
    })

    app.Get("/todo", func(c *fiber.Ctx) error {
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
    req := httptest.NewRequest("GET", "/appliances", nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }

    // add appliance
    body := models.Appliance{ApplianceName: "Xfridge", Manufacturer: "Acme"}
    b, _ := json.Marshal(body)
    req = httptest.NewRequest("POST", "/appliances/add", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, _ = app.Test(req)
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200 on add, got %d", resp.StatusCode)
    }

    // list should now have one
    req = httptest.NewRequest("GET", "/appliances", nil)
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
    req := httptest.NewRequest("POST", "/todo/add", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200 on add todo, got %d", resp.StatusCode)
    }

    // get todos
    req = httptest.NewRequest("GET", "/todo", nil)
    resp, _ = app.Test(req)
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200 on get todos, got %d", resp.StatusCode)
    }
}
