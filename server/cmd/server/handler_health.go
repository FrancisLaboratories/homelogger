package main

import (
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
	"github.com/masoncfrancis/homelogger/server/internal/version"
	"gorm.io/gorm"
)

func HealthHandler(db func() *gorm.DB, demoMode bool, importing *atomic.Bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		dbStatus := "ok"
		if dbConn := db(); dbConn == nil {
			dbStatus = "error: no database connection"
		} else if dbSQL, err := dbConn.DB(); err != nil {
			dbStatus = "error: " + err.Error()
		} else if err := dbSQL.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}

		status := fiber.Map{
			"status":    "ok",
			"version":   version.Version,
			"db":        dbStatus,
			"demo":      demoMode,
			"importing": importing.Load(),
		}

		if dbStatus != "ok" {
			return c.Status(fiber.StatusInternalServerError).JSON(status)
		}

		return c.Status(fiber.StatusOK).JSON(status)
	}
}
