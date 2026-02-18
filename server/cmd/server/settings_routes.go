package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"gorm.io/gorm"
)

func registerSettingsRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/settings", func(c *fiber.Ctx) error {
		settings, err := database.GetSettings(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting settings: " + err.Error())
		}
		return c.JSON(settings)
	})

	app.Put("/settings", func(c *fiber.Ctx) error {
		var body struct {
			Locale            *string `json:"locale"`
			Language          *string `json:"language"`
			Currency          *string `json:"currency"`
			TimeZone          *string `json:"timeZone"`
			MeasurementSystem *string `json:"measurementSystem"`
			WeekStart         *int    `json:"weekStart"`
			DateFormat        *string `json:"dateFormat"`
			NumberingSystem   *string `json:"numberingSystem"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}

		settings, err := database.GetSettings(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting settings: " + err.Error())
		}

		if body.Locale != nil {
			settings.Locale = *body.Locale
		}
		if body.Language != nil {
			settings.Language = *body.Language
		}
		if body.Currency != nil {
			settings.Currency = *body.Currency
		}
		if body.TimeZone != nil {
			settings.TimeZone = *body.TimeZone
		}
		if body.MeasurementSystem != nil {
			if *body.MeasurementSystem != "imperial" && *body.MeasurementSystem != "metric" {
				return c.Status(fiber.StatusBadRequest).SendString("measurementSystem must be 'imperial' or 'metric'")
			}
			settings.MeasurementSystem = *body.MeasurementSystem
		}
		if body.WeekStart != nil {
			if *body.WeekStart < 0 || *body.WeekStart > 6 {
				return c.Status(fiber.StatusBadRequest).SendString("weekStart must be between 0 and 6")
			}
			settings.WeekStart = *body.WeekStart
		}
		if body.DateFormat != nil {
			settings.DateFormat = *body.DateFormat
		}
		if body.NumberingSystem != nil {
			settings.NumberingSystem = *body.NumberingSystem
		}

		updated, err := database.UpdateSettings(db, settings)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating settings: " + err.Error())
		}
		return c.JSON(updated)
	})

	app.Get("/settings/options", func(c *fiber.Ctx) error {
		options := fiber.Map{
			"locales":            []string{"en-US", "en-ZA", "en-UK"},
			"languages":          []string{"en"},
			"currencies":         []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "ZAR"},
			"timeZones":          []string{"UTC", "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", "Europe/London", "Europe/Berlin", "Asia/Tokyo", "Australia/Sydney"},
			"measurementSystems": []string{"imperial", "metric"},
			"weekStartOptions":   []int{0, 1, 6},
			"dateFormats":        []string{"YYYY-MM-DD", "DD/MM/YYYY", "MM/DD/YYYY", "YYYY/MM/DD", "DD-MM-YYYY", "MM-DD-YYYY"},
			"numberingSystems":   []string{"latn"},
		}
		return c.JSON(options)
	})
}
