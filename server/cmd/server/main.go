package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/demo"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"github.com/masoncfrancis/homelogger/server/internal/version"
)

var backupMu sync.Mutex
var demoMu sync.Mutex

func main() {
	// CLI flags
	showVersion := flag.Bool("version", false, "Print version and exit")
	shortV := flag.Bool("v", false, "Print version and exit (shorthand)")
	flag.Parse()
	if (showVersion != nil && *showVersion) || (shortV != nil && *shortV) {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// Demo DB handling: if demo mode is enabled, use a separate demo DB file
	demoMode := false
	demoDBPath := ""
	if dm := os.Getenv("DEMO_MODE"); dm == "true" || dm == "1" {
		demoMode = true
		demoDBPath = "./data/db/demo.db"
		_ = os.Setenv("DEMO_DB_PATH", demoDBPath)
	}

	// Connect to GORM
	db, err := database.ConnectGorm()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Migrate GORM
	err = database.MigrateGorm(db)
	if err != nil {
		panic("Error migrating GORM")
	}

	if err := database.MigrateTodosToTasks(db); err != nil {
		fmt.Printf("Warning: todo→task migration failed: %v\n", err)
	}

	if demoMode && db != nil && db.Dialector.Name() == "postgres" {
		fmt.Println("Warning: DEMO_MODE is only supported with SQLite; disabling demo mode for PostgreSQL")
		demoMode = false
	}

	// Demo mode: optionally seed the DB from sample JSON when enabled.
	if demoMode {
		demoPath := os.Getenv("DEMO_FILE_PATH")
		if err := demo.Seed(db, demoPath); err != nil {
			fmt.Printf("Error seeding demo data: %v\n", err)
		}
		// record initial demo seed time
		_ = os.MkdirAll("./data", 0755)
		_ = os.WriteFile("./data/demo_last_reset", []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0644)
	}

	// If demo mode, start a background checker that resets demo data every 10 minutes.
	if demoMode {

		resetDemo := func() error {
			demoMu.Lock()
			defer demoMu.Unlock()

			demoPath := os.Getenv("DEMO_FILE_PATH")

			var errs []string

			// Close existing DB connection before replacing the file
			if db != nil {
				if sqlDB, err := db.DB(); err == nil {
					if err2 := sqlDB.Close(); err2 != nil {
						errs = append(errs, fmt.Sprintf("close db: %v", err2))
					}
				}
			}

			// Remove demo DB file
			if demoDBPath != "" {
				if err := os.Remove(demoDBPath); err != nil && !os.IsNotExist(err) {
					errs = append(errs, fmt.Sprintf("remove demo db: %v", err))
				}
			}

			// Remove demo uploads folder
			demoUploadsPath := filepath.Join("./data/uploads", "demo-uploads")
			if err := os.RemoveAll(demoUploadsPath); err != nil {
				errs = append(errs, fmt.Sprintf("remove demo uploads: %v", err))
			}

			// Reconnect and re-seed
			newDB, err := database.ConnectGorm()
			if err != nil {
				errs = append(errs, fmt.Sprintf("connect gorm: %v", err))
				return errors.New(strings.Join(errs, "; "))
			}
			db = newDB
			if err := database.MigrateGorm(db); err != nil {
				errs = append(errs, fmt.Sprintf("migrate gorm: %v", err))
				return errors.New(strings.Join(errs, "; "))
			}
			if err := demo.Seed(db, demoPath); err != nil {
				errs = append(errs, fmt.Sprintf("seed demo: %v", err))
				return errors.New(strings.Join(errs, "; "))
			}

			// update timestamp file
			if err := os.WriteFile("./data/demo_last_reset", []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0644); err != nil {
				errs = append(errs, fmt.Sprintf("write timestamp: %v", err))
			}

			if len(errs) > 0 {
				return errors.New(strings.Join(errs, "; "))
			}
			return nil
		}

		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				data, err := os.ReadFile("./data/demo_last_reset")
				var last int64 = 0
				if err == nil {
					if v, err2 := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err2 == nil {
						last = v
					}
				}
				elapsed := time.Now().Unix() - last
				minutesLeft := 10 - int(elapsed/60)
				if minutesLeft > 0 {
					fmt.Printf("Demo reset in %d minute(s)\n", minutesLeft)
					continue
				}
				fmt.Printf("Demo reset triggered now\n")
				if err := resetDemo(); err != nil {
					fmt.Printf("Demo reset failed: %v\n", err)
				} else {
					fmt.Printf("Demo reset completed successfully\n")
				}
			}
		}()
	}

	// Create new fiber server with larger body limit for file uploads
	app := fiber.New(fiber.Config{
		AppName:   fmt.Sprintf("HomeLogger %s", version.Version),
		BodyLimit: 100 * 1024 * 1024, // 100 MB
	})

	app.Hooks().OnPreStartupMessage(func(sm *fiber.PreStartupMessageData) error {
        sm.BannerHeader = "    __  __                     __                               \n" +
		"   / / / /___  ____ ___  ___  / /   ____  ____ _____ ____  _____\n" +
		"  / /_/ / __ \\/ __ `__ \\/ _ \\/ /   / __ \\/ __ `/ __ `/ _ \\/ ___/\n" +
		" / __  / /_/ / / / / / /  __/ /___/ /_/ / /_/ / /_/ /  __/ /    \n" +
		"/_/ /_/\\____/_/ /_/ /_/\\___/_____/\\____/\\__, /\\__, /\\___/_/     \n" +
		"                                       /____//____/             \n\n"

        return nil
    })

	// Use CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Allow all origins
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Request logging middleware
	logWriter := newLogWriter()
	app.Use(requestLogger(logWriter))

	// API routes grouped under /api
	api := app.Group("/api")

	// Health endpoint
	api.Get("/health", func(c fiber.Ctx) error {
		// Check DB connectivity
		dbSQL, err := db.DB()
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error: " + err.Error()
		} else if err := dbSQL.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}

		status := fiber.Map{
			"status":  "ok",
			"version": version.Version,
			"db":      dbStatus,
			"demo":    demoMode,
		}

		// If DB is not ok, return 500
		if dbStatus != "ok" {
			return c.Status(fiber.StatusInternalServerError).JSON(status)
		}

		return c.Status(fiber.StatusOK).JSON(status)
	})

	// Get all appliances
	api.Get("/appliances", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get all appliances
		appliances, err := database.GetAppliances(db)
		if err != nil {
			return c.SendString("Error getting appliances:" + err.Error())
		}

		return c.JSON(appliances)
	})

	// Create a new appliance
	api.Post("/appliances/add", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the appliance details from the body
		var body struct {
			ApplianceName string `json:"applianceName"`
			Manufacturer  string `json:"manufacturer"`
			ModelNumber   string `json:"modelNumber"`
			SerialNumber  string `json:"serialNumber"`
			YearPurchased string `json:"yearPurchased"`
			PurchasePrice string `json:"purchasePrice"`
			Location      string `json:"location"`
			Type          string `json:"type"`
		}
		err = c.Bind().Body(&body)
		if err != nil {
			return c.SendString("Error parsing body")
		}

		// Add an appliance
		appliance, err := database.AddAppliance(db, &models.Appliance{
			ApplianceName: body.ApplianceName,
			Manufacturer:  body.Manufacturer,
			ModelNumber:   body.ModelNumber,
			SerialNumber:  body.SerialNumber,
			YearPurchased: body.YearPurchased,
			PurchasePrice: body.PurchasePrice,
			Location:      body.Location,
			Type:          body.Type,
		})
		if err != nil {
			return c.SendString("Error adding appliance:" + err.Error())
		}

		return c.JSON(appliance)
	})

	// Update an appliance
	api.Put("/appliances/update/:id", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the id from the URL
		id := c.Params("id")

		// Convert id to uint
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.SendString("Invalid ID format")
		}

		// Get the appliance details from the body
		var body struct {
			ApplianceName string `json:"applianceName"`
			Manufacturer  string `json:"manufacturer"`
			ModelNumber   string `json:"modelNumber"`
			SerialNumber  string `json:"serialNumber"`
			YearPurchased string `json:"yearPurchased"`
			PurchasePrice string `json:"purchasePrice"`
			Location      string `json:"location"`
			Type          string `json:"type"`
		}
		err = c.Bind().Body(&body)
		if err != nil {
			return c.SendString("Error parsing body")
		}

		// Get the existing appliance
		appliance, err := database.GetAppliance(db, uint(idUint))
		if err != nil {
			return c.SendString("Error getting appliance:" + err.Error())
		}

		// Update the appliance details
		appliance.ApplianceName = body.ApplianceName
		appliance.Manufacturer = body.Manufacturer
		appliance.ModelNumber = body.ModelNumber
		appliance.SerialNumber = body.SerialNumber
		appliance.YearPurchased = body.YearPurchased
		appliance.PurchasePrice = body.PurchasePrice
		appliance.Location = body.Location
		appliance.Type = body.Type

		// Save the updated appliance
		updatedAppliance, err := database.UpdateAppliance(db, appliance)
		if err != nil {
			return c.SendString("Error updating appliance:" + err.Error())
		}

		return c.JSON(updatedAppliance)
	})

	api.Get("/appliances/:id", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the id from the URL
		id := c.Params("id")

		// Convert id to uint
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.SendString("Invalid ID format")
		}

		// Get the appliance
		appliance, err := database.GetAppliance(db, uint(idUint))
		if err != nil {
			return c.SendString("Error getting appliance:" + err.Error())
		}

		return c.JSON(appliance)
	})

	api.Delete("/appliances/delete/:id", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the id from the URL
		id := c.Params("id")

		// Convert id to uint
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.SendString("Invalid ID format")
		}

		// Delete the appliance
		err = database.DeleteAppliance(db, uint(idUint))
		if err != nil {
			return c.SendString("Error deleting appliance:" + err.Error())
		}

		return c.SendString("Appliance deleted")
	})

	// Maintenance endpoints
	api.Get("/maintenance", func(c fiber.Ctx) error {
		applianceId := c.Query("applianceId")
		referenceType := c.Query("referenceType")
		spaceType := c.Query("spaceType")

		if referenceType == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: referenceType")
		}

		if referenceType == "Space" {
			if spaceType == "" {
				return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: spaceType for Space reference")
			}
			maintenances, err := database.GetMaintenances(db, 0, referenceType, spaceType)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error getting maintenance records: " + err.Error())
			}
			return c.JSON(maintenances)
		}

		if applianceId == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: applianceId for Appliance reference")
		}

		applianceIdUint, err := strconv.ParseUint(applianceId, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid applianceId format")
		}

		maintenances, err := database.GetMaintenances(db, uint(applianceIdUint), referenceType, spaceType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting maintenance records: " + err.Error())
		}
		return c.JSON(maintenances)
	})

	api.Post("/maintenance/add", func(c fiber.Ctx) error {
		// Expect maintenance fields plus optional attachmentIds array
		var body struct {
			models.Maintenance
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		// Create maintenance record
		newMaintenance, err := database.AddMaintenance(db, &body.Maintenance)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error adding maintenance record: " + err.Error())
		}

		// Attach files if any
		for _, fid := range body.AttachmentIDs {
			_ = database.AttachFileToMaintenance(db, fid, newMaintenance.ID)
		}

		return c.Status(fiber.StatusCreated).JSON(newMaintenance)
	})

	api.Get("/maintenance/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		maintenance, err := database.GetMaintenance(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Maintenance record not found: " + err.Error())
		}
		return c.JSON(maintenance)
	})

	api.Delete("/maintenance/delete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		if err := database.DeleteMaintenance(db, uint(idUint)); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting maintenance record: " + err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	api.Put("/maintenance/update/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		var body struct {
			Description string  `json:"description"`
			Date        string  `json:"date"`
			Cost        float64 `json:"cost"`
			Notes       string  `json:"notes"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		updated, err := database.UpdateMaintenance(db, uint(idUint), body.Description, body.Date, body.Cost, body.Notes)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating maintenance record: " + err.Error())
		}
		return c.JSON(updated)
	})

	// Repair endpoints
	api.Get("/repair", func(c fiber.Ctx) error {
		applianceId := c.Query("applianceId")
		referenceType := c.Query("referenceType")
		spaceType := c.Query("spaceType")

		if referenceType == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: referenceType")
		}

		if referenceType == "Space" {
			if spaceType == "" {
				return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: spaceType for Space reference")
			}
			repairs, err := database.GetRepairs(db, 0, referenceType, spaceType)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error getting repair records: " + err.Error())
			}
			return c.JSON(repairs)
		}

		if applianceId == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing required query parameter: applianceId for Appliance reference")
		}

		applianceIdUint, err := strconv.ParseUint(applianceId, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid applianceId format")
		}

		repairs, err := database.GetRepairs(db, uint(applianceIdUint), referenceType, spaceType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting repair records: " + err.Error())
		}
		return c.JSON(repairs)
	})

	api.Post("/repair/add", func(c fiber.Ctx) error {
		var body struct {
			models.Repair
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		newRepair, err := database.AddRepair(db, &body.Repair)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error adding repair record: " + err.Error())
		}

		for _, fid := range body.AttachmentIDs {
			_ = database.AttachFileToRepair(db, fid, newRepair.ID)
		}

		return c.Status(fiber.StatusCreated).JSON(newRepair)
	})

	api.Get("/repair/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		repair, err := database.GetRepair(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Repair record not found: " + err.Error())
		}
		return c.JSON(repair)
	})

	api.Delete("/repair/delete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		if err := database.DeleteRepair(db, uint(idUint)); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting repair record: " + err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	api.Put("/repair/update/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		var body struct {
			Description string  `json:"description"`
			Date        string  `json:"date"`
			Cost        float64 `json:"cost"`
			Notes       string  `json:"notes"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		updated, err := database.UpdateRepair(db, uint(idUint), body.Description, body.Date, body.Cost, body.Notes)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating repair record: " + err.Error())
		}
		return c.JSON(updated)
	})

	// Upload a new file
	api.Post("/files/upload", func(c fiber.Ctx) error {
		// Parse the multipart form
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing multipart form: " + err.Error())
		}

		// Get the file and userID from the form
		files := form.File["file"]
		userID := ""
		if vals, ok := form.Value["userID"]; ok && len(vals) > 0 {
			userID = vals[0]
		}

		// optional spaceType
		spaceType := ""
		if vals, ok := form.Value["spaceType"]; ok && len(vals) > 0 {
			spaceType = vals[0]
		}

		if len(files) == 0 || userID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing file or userID")
		}

		// Create a new SavedFile object without setting the ID
		file := files[0]
		savedFile := &models.SavedFile{
			OriginalName: file.Filename,
			Type:         "",
			UserID:       userID,
		}

		if spaceType != "" {
			savedFile.SpaceType = &spaceType
		}

		// Save the file information to the database
		newFile, err := database.UploadFile(db, savedFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error saving file information: " + err.Error())
		}

		// Save the file to the server with the id as the file name.
		// If demo mode is enabled, save under a demo-uploads subfolder.
		uploadsBase := "./data/uploads"
		if demoMode {
			uploadsBase = filepath.Join(uploadsBase, "demo-uploads")
		}
		// Ensure the uploads directory exists
		if err := os.MkdirAll(uploadsBase, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error creating uploads directory: " + err.Error())
		}
		filePath := filepath.Join(uploadsBase, strconv.FormatUint(uint64(newFile.ID), 10))
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error saving file: " + err.Error())
		}

		// Update the file path in the database
		newFile.Path = filePath
		if _, err := database.UpdateFilePath(db, newFile); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating file path: " + err.Error())
		}

		// Return the id, originalName, and userID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":           newFile.ID,
			"originalName": newFile.OriginalName,
			"userID":       newFile.UserID,
		})
	})

	// Get file information by ID
	api.Get("/files/info/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		fileInfo, err := database.GetFileInfo(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("File not found: " + err.Error())
		}

		return c.JSON(fileInfo)
	})

	// List files attached to a maintenance record
	api.Get("/files/maintenance/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		files, err := database.GetFilesByMaintenance(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting files: " + err.Error())
		}
		return c.JSON(files)
	})

	// List files attached to a repair record
	api.Get("/files/repair/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		files, err := database.GetFilesByRepair(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting files: " + err.Error())
		}
		return c.JSON(files)
	})

	api.Get("/files/download/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		// Fetch the file information
		fileInfo, err := database.GetFileInfo(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("File not found: " + err.Error())
		}

		// Fetch the file path using the GetFilePath function
		filePath, err := database.GetFilePath(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("File path not found: " + err.Error())
		}

		// Set the Content-Disposition header to specify the original file name
		c.Set("Content-Disposition", "attachment; filename="+fileInfo.OriginalName)

		return c.SendFile(filePath)
	})

	// List files attached to an appliance
	api.Get("/files/appliance/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		files, err := database.GetFilesByAppliance(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting files: " + err.Error())
		}
		return c.JSON(files)
	})

	// List files attached to a space type
	api.Get("/files/space/:spaceType", func(c fiber.Ctx) error {
		spaceType := c.Params("spaceType")
		if spaceType == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing spaceType")
		}

		files, err := database.GetFilesBySpace(db, spaceType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting files: " + err.Error())
		}
		return c.JSON(files)
	})

	// Notes endpoints
	api.Get("/notes", func(c fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get optional filters
		applianceIdStr := c.Query("applianceId")
		spaceType := c.Query("spaceType")
		var applianceId uint = 0
		if applianceIdStr != "" {
			if idUint, err := strconv.ParseUint(applianceIdStr, 10, 32); err == nil {
				applianceId = uint(idUint)
			}
		}

		notes, err := database.GetNotes(db, applianceId, spaceType)
		if err != nil {
			return c.SendString("Error getting notes:" + err.Error())
		}

		return c.JSON(notes)
	})

	api.Post("/notes/add", func(c fiber.Ctx) error {
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		var body struct {
			Title       string `json:"title"`
			Body        string `json:"body"`
			ApplianceID uint   `json:"applianceId"`
			SpaceType   string `json:"spaceType"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.SendString("Error parsing body")
		}

		var applianceId uint = 0
		if body.ApplianceID != 0 {
			applianceId = body.ApplianceID
		}

		note, err := database.AddNote(db, body.Title, body.Body, applianceId, body.SpaceType)
		if err != nil {
			return c.SendString("Error adding note:" + err.Error())
		}

		return c.JSON(note)
	})

	api.Get("/notes/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		note, err := database.GetNote(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Note not found: " + err.Error())
		}
		return c.JSON(note)
	})

	api.Put("/notes/update/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		var body struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.SendString("Error parsing body")
		}

		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		updated, err := database.UpdateNote(db, uint(idUint), body.Title, body.Body)
		if err != nil {
			return c.SendString("Error updating note:" + err.Error())
		}

		return c.JSON(updated)
	})

	api.Delete("/notes/delete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.SendString("Invalid ID format")
		}

		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		err = database.DeleteNote(db, uint(idUint))
		if err != nil {
			return c.SendString("Error deleting note:" + err.Error())
		}

		return c.SendString("Note deleted")
	})

	// Associate an existing uploaded file with a maintenance, repair, appliance, or space
	api.Post("/files/attach", func(c fiber.Ctx) error {
		var body struct {
			FileID        uint   `json:"fileId"`
			MaintenanceID uint   `json:"maintenanceId"`
			RepairID      uint   `json:"repairId"`
			ApplianceID   uint   `json:"applianceId"`
			SpaceType     string `json:"spaceType"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}

		if body.MaintenanceID != 0 {
			if err := database.AttachFileToMaintenance(db, body.FileID, body.MaintenanceID); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error attaching file: " + err.Error())
			}
		}
		if body.RepairID != 0 {
			if err := database.AttachFileToRepair(db, body.FileID, body.RepairID); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error attaching file: " + err.Error())
			}
		}
		if body.ApplianceID != 0 {
			if err := database.AttachFileToAppliance(db, body.FileID, body.ApplianceID); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error attaching file: " + err.Error())
			}
		}

		if body.SpaceType != "" {
			if err := database.AttachFileToSpace(db, body.FileID, body.SpaceType); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error attaching file to space: " + err.Error())
			}
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Delete a file (record + stored file)
	api.Delete("/files/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		// Get file path
		filePath, err := database.GetFilePath(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("File path not found: " + err.Error())
		}

		// Delete file from disk if exists
		if err := os.Remove(filePath); err != nil {
			// If file doesn't exist, continue to delete DB record
		}

		// Delete DB record
		if err := database.DeleteFile(db, uint(idUint)); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting file record: " + err.Error())
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Task endpoints
	api.Get("/task", func(c fiber.Ctx) error {
		applianceIdStr := c.Query("applianceId")
		spaceType := c.Query("spaceType")
		includeCompleted := c.Query("includeCompleted") == "true"

		var applianceId uint = 0
		if applianceIdStr != "" {
			if idUint, err := strconv.ParseUint(applianceIdStr, 10, 32); err == nil {
				applianceId = uint(idUint)
			}
		}

		tasks, err := database.GetTasks(db, applianceId, spaceType, includeCompleted)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting tasks: " + err.Error())
		}
		return c.JSON(tasks)
	})

	api.Get("/task/dashboard", func(c fiber.Ctx) error {
		includeCompleted := fiber.Query[bool](c, "includeCompleted", false)
		tasks, err := database.GetAllTasks(db, includeCompleted)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error getting tasks: " + err.Error())
		}
		return c.JSON(tasks)
	})

	api.Post("/task/add", func(c fiber.Ctx) error {
		var body struct {
			Label              string   `json:"label"`
			Notes              string   `json:"notes"`
			Priority           string   `json:"priority"`
			DueDate            *string  `json:"dueDate"`
			EstimatedCost      *float64 `json:"estimatedCost"`
			IsRecurring        bool     `json:"isRecurring"`
			RecurrenceInterval int      `json:"recurrenceInterval"`
			RecurrenceUnit     string   `json:"recurrenceUnit"`
			RecurrenceMode     string   `json:"recurrenceMode"`
			ApplianceID        *uint    `json:"applianceId"`
			SpaceType          *string  `json:"spaceType"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		if body.Label == "" {
			return c.Status(fiber.StatusBadRequest).SendString("label is required")
		}

		task := &models.Task{
			Label:              body.Label,
			Notes:              body.Notes,
			Priority:           body.Priority,
			DueDate:            body.DueDate,
			EstimatedCost:      body.EstimatedCost,
			IsRecurring:        body.IsRecurring,
			RecurrenceInterval: body.RecurrenceInterval,
			RecurrenceUnit:     body.RecurrenceUnit,
			RecurrenceMode:     body.RecurrenceMode,
			ApplianceID:        body.ApplianceID,
			SpaceType:          body.SpaceType,
			UserID:             "1",
		}

		created, err := database.AddTask(db, task)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error adding task: " + err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(created)
	})

	api.Get("/task/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		task, err := database.GetTask(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Task not found: " + err.Error())
		}
		return c.JSON(task)
	})

	api.Put("/task/update/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		existing, err := database.GetTask(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Task not found: " + err.Error())
		}

		var body struct {
			Label              string   `json:"label"`
			Notes              string   `json:"notes"`
			Priority           string   `json:"priority"`
			DueDate            *string  `json:"dueDate"`
			EstimatedCost      *float64 `json:"estimatedCost"`
			IsRecurring        bool     `json:"isRecurring"`
			RecurrenceInterval int      `json:"recurrenceInterval"`
			RecurrenceUnit     string   `json:"recurrenceUnit"`
			RecurrenceMode     string   `json:"recurrenceMode"`
			ApplianceID        *uint    `json:"applianceId"`
			SpaceType          *string  `json:"spaceType"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}

		existing.Label = body.Label
		existing.Notes = body.Notes
		existing.Priority = body.Priority
		existing.DueDate = body.DueDate
		existing.EstimatedCost = body.EstimatedCost
		existing.IsRecurring = body.IsRecurring
		existing.RecurrenceInterval = body.RecurrenceInterval
		existing.RecurrenceUnit = body.RecurrenceUnit
		existing.RecurrenceMode = body.RecurrenceMode
		existing.ApplianceID = body.ApplianceID
		existing.SpaceType = body.SpaceType

		updated, err := database.UpdateTask(db, existing)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating task: " + err.Error())
		}
		return c.JSON(updated)
	})

	api.Put("/task/complete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}

		var body struct {
			CompletionDate string  `json:"completionDate"`
			CreateRecord   bool    `json:"createRecord"`
			RecordType     string  `json:"recordType"`
			Description    string  `json:"description"`
			Cost           float64 `json:"cost"`
		}
		if err := c.Bind().Body(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing body: " + err.Error())
		}
		if body.CompletionDate == "" {
			return c.Status(fiber.StatusBadRequest).SendString("completionDate is required")
		}

		task, err := database.CompleteTask(db, uint(idUint), body.CompletionDate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error completing task: " + err.Error())
		}

		// Optionally create a Maintenance or Repair record
		if body.CreateRecord {
			description := body.Description
			if description == "" {
				description = task.Label
			}

			// Determine reference type from the task
			refType := "Space"
			spaceType := ""
			var applianceId *uint
			if task.ApplianceID != nil {
				refType = "Appliance"
				applianceId = task.ApplianceID
			} else if task.SpaceType != nil {
				spaceType = *task.SpaceType
			}

			if body.RecordType == "repair" {
				repair := &models.Repair{
					Description:   description,
					Date:          body.CompletionDate,
					Cost:          body.Cost,
					Notes:         "",
					SpaceType:     spaceType,
					ReferenceType: refType,
					ApplianceID:   applianceId,
				}
				if _, err := database.AddRepair(db, repair); err != nil {
					return c.Status(fiber.StatusInternalServerError).SendString("Error creating repair record: " + err.Error())
				}
			} else {
				maintenance := &models.Maintenance{
					Description:   description,
					Date:          body.CompletionDate,
					Cost:          body.Cost,
					Notes:         "",
					SpaceType:     spaceType,
					ReferenceType: refType,
					ApplianceID:   applianceId,
				}
				if _, err := database.AddMaintenance(db, maintenance); err != nil {
					return c.Status(fiber.StatusInternalServerError).SendString("Error creating maintenance record: " + err.Error())
				}
			}
		}

		return c.JSON(task)
	})

	api.Put("/task/uncomplete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		task, err := database.UncompleteTask(db, uint(idUint))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error uncompleting task: " + err.Error())
		}
		return c.JSON(task)
	})

	api.Delete("/task/delete/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
		}
		if err := database.DeleteTask(db, uint(idUint)); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting task: " + err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Download a backup ZIP containing the DB and uploads
	api.Get("/backup/download", func(c fiber.Ctx) error {
		pr, pw := io.Pipe()

		go func() {
			zw := zip.NewWriter(pw)
			defer func() {
				_ = zw.Close()
				_ = pw.Close()
			}()

			backupMu.Lock()
			defer backupMu.Unlock()

			// ponytail: Universal JSON export — works on any GORM dialect, no raw dump needed.
			payload, err := database.ExportToJSON(db, db.Dialector.Name())
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("export data: %w", err))
				return
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("marshal payload: %w", err))
				return
			}

			w, err := zw.Create("data.json")
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("zip entry data.json: %w", err))
				return
			}
			if _, err := w.Write(jsonData); err != nil {
				_ = pw.CloseWithError(fmt.Errorf("write data.json: %w", err))
				return
			}

			uploadsRoot := "./data/uploads"
			_ = filepath.Walk(uploadsRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				rel, err := filepath.Rel(uploadsRoot, path)
				if err != nil {
					return err
				}
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				dst, err := zw.Create(filepath.ToSlash(filepath.Join("uploads", rel)))
				if err != nil {
					return err
				}
				_, err = io.Copy(dst, f)
				return err
			})
		}()

		c.Set("Content-Type", "application/zip")
		c.Set("Content-Disposition", "attachment; filename=homelogger-backup.zip")
		return c.SendStream(pr)
	})

	// Import a backup ZIP — replaces all data: drop tables → migrate → insert
	api.Post("/backup/import", func(c fiber.Ctx) error {
		backupMu.Lock()
		defer backupMu.Unlock()

		file, err := c.FormFile("backup")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error getting backup file: " + err.Error())
		}

		tempDir, err := os.MkdirTemp("", "homelogger-backup-import-")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error creating temp directory: " + err.Error())
		}
		defer func() { _ = os.RemoveAll(tempDir) }()

		tempZipPath := filepath.Join(tempDir, file.Filename)
		if err := c.SaveFile(file, tempZipPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error saving uploaded file: " + err.Error())
		}

		r, err := zip.OpenReader(tempZipPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error opening zip file: " + err.Error())
		}
		defer func() { _ = r.Close() }()

		extractedPath := filepath.Join(tempDir, "extracted")
		if err := os.MkdirAll(extractedPath, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error creating extraction directory: " + err.Error())
		}

		var dataJSONPath string
		var legacyDBPath string
		var uploadsExtractedPath string

		for _, f := range r.File {
			fpath := filepath.Join(extractedPath, f.Name)
			if !strings.HasPrefix(fpath, filepath.Clean(extractedPath)+string(os.PathSeparator)) {
				return c.Status(fiber.StatusBadRequest).SendString("Illegal file path in zip: " + fpath)
			}
			if f.FileInfo().IsDir() {
				_ = os.MkdirAll(fpath, os.ModePerm)
				continue
			}
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error creating dir: " + err.Error())
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error creating file: " + err.Error())
			}
			rc, err := f.Open()
			if err != nil {
				_ = outFile.Close()
				return c.Status(fiber.StatusInternalServerError).SendString("Error opening zip entry: " + err.Error())
			}
			_, copyErr := io.Copy(outFile, rc)
			_ = outFile.Close()
			_ = rc.Close()
			if copyErr != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error extracting file: " + copyErr.Error())
			}
			if strings.EqualFold(filepath.Base(fpath), "data.json") {
				dataJSONPath = fpath
			} else if strings.HasPrefix(f.Name, "db/") && strings.HasSuffix(strings.ToLower(f.Name), ".db") && legacyDBPath == "" {
				legacyDBPath = fpath
			} else if strings.HasPrefix(f.Name, "uploads/") && uploadsExtractedPath == "" {
				uploadsExtractedPath = filepath.Join(extractedPath, "uploads")
			}
		}

		switch {
		case dataJSONPath != "":
			if _, err := database.ImportFromJSONFile(db, dataJSONPath, uploadsExtractedPath); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error importing database data: " + err.Error())
			}
		case legacyDBPath != "":
			payload, err := database.ConvertLegacyDB(legacyDBPath)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error reading legacy backup: " + err.Error())
			}
			if _, err := database.ImportFromJSON(db, payload, uploadsExtractedPath); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error importing database data: " + err.Error())
			}
		default:
			return c.Status(fiber.StatusBadRequest).SendString("Backup ZIP must contain data.json (new format) or a .db file in a db/ directory (legacy format)")
		}

		if err := database.ImportUploads(uploadsExtractedPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error importing uploaded files: " + err.Error())
		}

		return c.SendString("Backup import completed successfully")
	})

	// Serve static SPA files with client-side routing fallback
	app.Get("/*", static.New("./static"), func(c fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":3005"
	}
	fmt.Printf("\n\nStarting HomeLogger %s on port %s\n\n", version.Version, addr)

	// Start server in goroutine so we can handle signals and cleanup
	serverErr := make(chan error, 1)
	go func() {
		if err := app.Listen(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for termination signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		fmt.Printf("Received signal %v, shutting down\n", sig)
	case err := <-serverErr:
		fmt.Printf("Server error: %v\n", err)
	}

	// Attempt graceful shutdown
	if err := app.Shutdown(); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
	}

	// Close log file
	logWriter.Close()

	// Close DB connection
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}

	// Remove demo DB file if demo mode
	if demoMode && demoDBPath != "" {
		if err := os.Remove(demoDBPath); err != nil {
			fmt.Printf("Error removing demo DB %s: %v\n", demoDBPath, err)
		} else {
			fmt.Printf("Removed demo DB %s\n", demoDBPath)
		}
	}

	// Remove demo uploads folder if demo mode
	if demoMode {
		demoUploadsPath := filepath.Join("./data/uploads", "demo-uploads")
		if err := os.RemoveAll(demoUploadsPath); err != nil {
			fmt.Printf("Error removing demo uploads %s: %v\n", demoUploadsPath, err)
		} else {
			fmt.Printf("Removed demo uploads %s\n", demoUploadsPath)
		}
	}
}
