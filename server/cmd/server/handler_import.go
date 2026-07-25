package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/masoncfrancis/homelogger/server/internal/database"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

func ImportBackupHandler(db *gorm.DB, importing *atomic.Bool, backupMu *sync.Mutex) fiber.Handler {
	return func(c fiber.Ctx) error {
		backupMu.Lock()
		defer backupMu.Unlock()

		importing.Store(true)
		defer importing.Store(false)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

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
		var nestedDataJSON string
		var nestedLegacyDB string
		var nestedUploads string

		for _, f := range r.File {
			if err := ctx.Err(); err != nil {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"status": "failed",
					"error":  "Import timed out during ZIP extraction",
				})
			}
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
			switch {
			case f.Name == "data.json":
				dataJSONPath = fpath
			case strings.HasSuffix(f.Name, "/data.json") && nestedDataJSON == "":
				nestedDataJSON = f.Name
			}
			if strings.HasPrefix(f.Name, "db/") && strings.HasSuffix(strings.ToLower(f.Name), ".db") && legacyDBPath == "" {
				legacyDBPath = fpath
			} else if strings.Contains(f.Name, "/db/") && strings.HasSuffix(strings.ToLower(f.Name), ".db") && nestedLegacyDB == "" {
				nestedLegacyDB = f.Name
			}
			if strings.HasPrefix(f.Name, "uploads/") && uploadsExtractedPath == "" {
				uploadsExtractedPath = filepath.Join(extractedPath, "uploads")
			} else if strings.Contains(f.Name, "/uploads/") && nestedUploads == "" {
				nestedUploads = f.Name
			}
		}

		if err := ctx.Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "failed",
				"error":  "Import timed out before database import",
			})
		}

		if dataJSONPath == "" && nestedDataJSON != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "failed",
				"error":  fmt.Sprintf("data.json was found inside %q — place it at the root of the ZIP archive", nestedDataJSON),
			})
		}
		if dataJSONPath != "" && uploadsExtractedPath == "" && nestedUploads != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "failed",
				"error":  fmt.Sprintf("uploads folder was found inside %q — place uploads/ at the root of the ZIP archive", nestedUploads),
			})
		}

		dbCtx := db.WithContext(ctx)
		var importResult *models.ImportResult
		switch {
		case dataJSONPath != "":
			importResult, err = database.ImportFromJSONFile(dbCtx, dataJSONPath, uploadsExtractedPath)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":   "failed",
					"importId": importResult.GetImportID(),
					"error":    "Error importing database data: " + err.Error(),
				})
			}
		case legacyDBPath != "":
			payload, convErr := database.ConvertLegacyDB(legacyDBPath)
			if convErr != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status": "failed",
					"error":  "Error reading legacy backup: " + convErr.Error(),
				})
			}
			importResult, err = database.ImportFromJSON(dbCtx, payload, uploadsExtractedPath)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":   "failed",
					"importId": importResult.GetImportID(),
					"error":    "Error importing database data: " + err.Error(),
				})
			}
		default:
			msg := "Backup ZIP must contain data.json (new format) or a .db file in a db/ directory (legacy format)"
			switch {
			case nestedDataJSON != "":
				msg += fmt.Sprintf(" data.json was found inside %q — place it at the root of the ZIP", nestedDataJSON)
			case nestedLegacyDB != "":
				msg += fmt.Sprintf(" legacy database was found inside %q — place it at the root of the ZIP", nestedLegacyDB)
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "failed",
				"error":  msg,
			})
		}

		if err := ctx.Err(); err != nil {
			database.FailImport(db, importResult.ImportID, "import timed out after database import")
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "failed",
				"importId": importResult.ImportID,
				"error":    "Import timed out after database import — uploads were not restored",
			})
		}

		if err := database.ImportUploads(uploadsExtractedPath); err != nil {
			database.FailImport(db, importResult.ImportID, err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":   "failed",
				"importId": importResult.ImportID,
				"inserted": importResult.Inserted,
				"error":    "Error importing uploaded files: " + err.Error(),
			})
		}

		database.CompleteImport(db, importResult.ImportID)
		return c.JSON(fiber.Map{
			"status":   "completed",
			"importId": importResult.ImportID,
			"inserted": importResult.Inserted,
		})
	}
}
