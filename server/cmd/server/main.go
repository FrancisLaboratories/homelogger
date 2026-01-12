package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"github.com/masoncfrancis/homelogger/server/internal/version"
)

var backupMu sync.Mutex

func main() {
	// CLI flags
	showVersion := flag.Bool("version", false, "Print version and exit")
	shortV := flag.Bool("v", false, "Print version and exit (shorthand)")
	flag.Parse()
	if (showVersion != nil && *showVersion) || (shortV != nil && *shortV) {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// Connect to GORM
	db, err := database.ConnectGorm()
	if err != nil {
		panic("Error connecting GORM to db")
	}

	// Migrate GORM
	err = database.MigrateGorm(db)
	if err != nil {
		panic("Error migrating GORM")
	}

	// Create new fiber server
	app := fiber.New()

	// Use CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow all origins
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World")
	})

	// Health endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
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
		}

		// If DB is not ok, return 500
		if dbStatus != "ok" {
			return c.Status(fiber.StatusInternalServerError).JSON(status)
		}

		return c.Status(fiber.StatusOK).JSON(status)
	})

	app.Get("/todo", func(c *fiber.Ctx) error {
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

		// Get todos with optional filters
		todos, err := database.GetTodos(db, applianceId, spaceType)
		if err != nil {
			return c.SendString("Error getting todos:" + err.Error())
		}

		return c.JSON(todos)
	})

	app.Put("/todo/update/:id", func(c *fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the id from the URL
		id := c.Params("id")

		// Get the checked status from the body
		var body struct {
			Checked bool `json:"checked"`
		}
		err = c.BodyParser(&body)
		if err != nil {
			return c.SendString("Error parsing body")
		}

		// Convert id to uint
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return c.SendString("Invalid ID format")
		}

		// Change the checked status of the todo
		err = database.ChangeTodoChecked(db, uint(idUint), body.Checked)
		if err != nil {
			return c.SendString("Error changing todo checked status:" + err.Error())
		}

		return c.SendString("Todo updated")
	})

	app.Post("/todo/add", func(c *fiber.Ctx) error {
		// Connect to gorm
		db, err := database.ConnectGorm()
		if err != nil {
			return c.SendString("Error connecting GORM to db")
		}

		// Get the label, checked status, and userid from the body
		var body struct {
			Label   string `json:"label"`
			Checked bool   `json:"checked"`
			UserID  string `json:"userid"`
		}
		err = c.BodyParser(&body)
		if err != nil {
			return c.SendString("Error parsing body")
		}

		// Add a todo (may include optional applianceId/spaceType)
		var applianceId uint = 0
		if bodyMap := c.Body(); len(bodyMap) > 0 {
			// body parsed into struct already; we'll parse optional fields separately below
		}
		// parse optional fields from a secondary struct
		var opt struct {
			ApplianceID uint   `json:"applianceId"`
			SpaceType   string `json:"spaceType"`
		}
		_ = c.BodyParser(&opt)

		if opt.ApplianceID != 0 {
			applianceId = opt.ApplianceID
		}

		todo, err := database.AddTodo(db, body.Label, body.Checked, body.UserID, applianceId, opt.SpaceType)
		if err != nil {
			return c.SendString("Error adding todo:" + err.Error())
		}

		return c.JSON(todo)
	})

	app.Delete("/todo/delete/:id", func(c *fiber.Ctx) error {
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

		// Delete the todo
		err = database.DeleteTodo(db, uint(idUint))
		if err != nil {
			return c.SendString("Error deleting todo:" + err.Error())
		}

		return c.SendString("Todo deleted")
	})

	// Get all appliances
	app.Get("/appliances", func(c *fiber.Ctx) error {
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
	app.Post("/appliances/add", func(c *fiber.Ctx) error {
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
		err = c.BodyParser(&body)
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
	app.Put("/appliances/update/:id", func(c *fiber.Ctx) error {
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
		err = c.BodyParser(&body)
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

	app.Get("/appliances/:id", func(c *fiber.Ctx) error {
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

	app.Delete("/appliances/delete/:id", func(c *fiber.Ctx) error {
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
	app.Get("/maintenance", func(c *fiber.Ctx) error {
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

	app.Post("/maintenance/add", func(c *fiber.Ctx) error {
		// Expect maintenance fields plus optional attachmentIds array
		var body struct {
			models.Maintenance
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.BodyParser(&body); err != nil {
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

	app.Get("/maintenance/:id", func(c *fiber.Ctx) error {
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

	app.Delete("/maintenance/delete/:id", func(c *fiber.Ctx) error {
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

	// Repair endpoints
	app.Get("/repair", func(c *fiber.Ctx) error {
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

	app.Post("/repair/add", func(c *fiber.Ctx) error {
		var body struct {
			models.Repair
			AttachmentIDs []uint `json:"attachmentIds"`
		}
		if err := c.BodyParser(&body); err != nil {
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

	app.Get("/repair/:id", func(c *fiber.Ctx) error {
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

	app.Delete("/repair/delete/:id", func(c *fiber.Ctx) error {
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

	// Upload a new file
	app.Post("/files/upload", func(c *fiber.Ctx) error {
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

		// Save the file to the server with the id as the file name
		filePath := "./data/uploads/" + strconv.FormatUint(uint64(newFile.ID), 10)
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
	app.Get("/files/info/:id", func(c *fiber.Ctx) error {
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
	app.Get("/files/maintenance/:id", func(c *fiber.Ctx) error {
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
	app.Get("/files/repair/:id", func(c *fiber.Ctx) error {
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

	app.Get("/files/download/:id", func(c *fiber.Ctx) error {
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
	app.Get("/files/appliance/:id", func(c *fiber.Ctx) error {
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
	app.Get("/files/space/:spaceType", func(c *fiber.Ctx) error {
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

	// Associate an existing uploaded file with a maintenance, repair, appliance, or space
	app.Post("/files/attach", func(c *fiber.Ctx) error {
		var body struct {
			FileID        uint   `json:"fileId"`
			MaintenanceID uint   `json:"maintenanceId"`
			RepairID      uint   `json:"repairId"`
			ApplianceID   uint   `json:"applianceId"`
			SpaceType     string `json:"spaceType"`
		}
		if err := c.BodyParser(&body); err != nil {
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
	app.Delete("/files/:id", func(c *fiber.Ctx) error {
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

	// Download a backup ZIP containing the DB and uploads
	app.Get("/backup/download", func(c *fiber.Ctx) error {
		pr, pw := io.Pipe()

		go func() {
			zw := zip.NewWriter(pw)
			// ensure the writer is closed which also closes the underlying pipe writer
			defer func() {
				_ = zw.Close()
				_ = pw.Close()
			}()

			// Create a consistent DB backup using the sqlite3 online backup API.
			dbPath := "./data/db/homelogger.db"
			tmpBackup := filepath.Join("./data/db", fmt.Sprintf("homelogger-backup-%d.db", time.Now().UnixNano()))

			// Copy fallback helper
			copyFile := func(src, dst string) error {
				in, err := os.Open(src)
				if err != nil {
					return err
				}
				defer in.Close()
				out, err := os.Create(dst)
				if err != nil {
					return err
				}
				defer out.Close()
				if _, err := io.Copy(out, in); err != nil {
					return err
				}
				return out.Sync()
			}

			// Serialize backups to avoid multiple concurrent backup attempts
			backupMu.Lock()
			defer backupMu.Unlock()

			// Attempt to create a consistent DB copy using SQLite's "VACUUM INTO" SQL
			// Use ExecContext with a short timeout so we don't block indefinitely.
			backedUp := false
			if db != nil {
				if dbSQL, err := db.DB(); err == nil {
					vacCtx, vacCancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer vacCancel()
					// Use VACUUM INTO which creates a consistent copy of the DB
					// Note: VACUUM INTO requires SQLite 3.27+
					vacuumSQL := fmt.Sprintf("VACUUM INTO '%s'", tmpBackup)
					if _, err := dbSQL.ExecContext(vacCtx, vacuumSQL); err == nil {
						backedUp = true
					}
				}
			}

			if !backedUp {
				// fallback to copying the file
				if err := copyFile(dbPath, tmpBackup); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}

			// Ensure tmpBackup is removed after adding
			defer func() {
				_ = os.Remove(tmpBackup)
			}()

			// Add the DB backup file into the ZIP
			if info, err := os.Stat(tmpBackup); err == nil && !info.IsDir() {
				f, err := os.Open(tmpBackup)
				if err != nil {
					_ = pw.CloseWithError(err)
					return
				}
				func() {
					defer f.Close()
					dst, err := zw.Create("db/" + filepath.Base(tmpBackup))
					if err != nil {
						_ = pw.CloseWithError(err)
						return
					}
					if _, err := io.Copy(dst, f); err != nil {
						_ = pw.CloseWithError(err)
						return
					}
				}()
			}

			// Walk uploads directory and add files preserving relative paths
			uploadsRoot := "./data/uploads"
			_ = filepath.Walk(uploadsRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
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

				// Use forward slashes inside ZIP
				dest := filepath.ToSlash(filepath.Join("uploads", rel))
				dst, err := zw.Create(dest)
				if err != nil {
					return err
				}
				if _, err := io.Copy(dst, f); err != nil {
					return err
				}
				return nil
			})
		}()

		c.Set("Content-Type", "application/zip")
		c.Set("Content-Disposition", "attachment; filename=homelogger-backup.zip")
		return c.SendStream(pr)
	})

	fmt.Printf("Starting HomeLogger Server %s on port 8083\n", version.Version)
	app.Listen(":8083")
}
