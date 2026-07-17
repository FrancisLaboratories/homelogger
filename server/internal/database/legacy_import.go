package database

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// ConvertLegacyDB opens a legacy SQLite backup file (from the old backup format)
// and reads all entity tables into a BackupPayload, compatible with ImportFromJSON.
// Tables that don't exist in the old DB (e.g. "tasks" in pre-migration backups)
// are skipped gracefully.
func ConvertLegacyDB(dbPath string) (*models.BackupPayload, error) {
	oldDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open legacy db: %w", err)
	}

	sqlDB, err := oldDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	payload := &models.BackupPayload{
		Version:      BackupVersion,
		ExportedAt:   time.Now().UTC(),
		DatabaseType: dialectSQLite,
	}

	migrator := oldDB.Migrator()

	if migrator.HasTable("appliances") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Appliances).Error; err != nil {
			return nil, fmt.Errorf("read appliances: %w", err)
		}
	}

	if migrator.HasTable("todos") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Todos).Error; err != nil {
			return nil, fmt.Errorf("read todos: %w", err)
		}
	}

	if migrator.HasTable("tasks") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Tasks).Error; err != nil {
			return nil, fmt.Errorf("read tasks: %w", err)
		}
	}

	if migrator.HasTable("maintenances") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Maintenance).Error; err != nil {
			return nil, fmt.Errorf("read maintenances: %w", err)
		}
	}

	if migrator.HasTable("repairs") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Repairs).Error; err != nil {
			return nil, fmt.Errorf("read repairs: %w", err)
		}
	}

	if migrator.HasTable("saved_files") {
		if err := oldDB.Unscoped().Find(&payload.Entities.SavedFiles).Error; err != nil {
			return nil, fmt.Errorf("read saved_files: %w", err)
		}
	}

	if migrator.HasTable("notes") {
		if err := oldDB.Unscoped().Find(&payload.Entities.Notes).Error; err != nil {
			return nil, fmt.Errorf("read notes: %w", err)
		}
	}

	// SQLite doesn't enforce FK constraints by default, so legacy backups may
	// have orphaned FK references (e.g. maintenance.appliance_id pointing to a
	// non-existent or hard-deleted appliance). NIL out any FK that doesn't
	// resolve to a record in the payload to avoid FK violations on Postgres.
	sanitizeFKs(payload)

	return payload, nil
}

func sanitizeFKs(payload *models.BackupPayload) {
	validApplianceIDs := make(map[uint]struct{}, len(payload.Entities.Appliances))
	for _, a := range payload.Entities.Appliances {
		validApplianceIDs[a.ID] = struct{}{}
	}

	validSavedFileIDs := make(map[uint]struct{}, len(payload.Entities.SavedFiles))
	for _, f := range payload.Entities.SavedFiles {
		validSavedFileIDs[f.ID] = struct{}{}
	}

	validMaintenanceIDs := make(map[uint]struct{}, len(payload.Entities.Maintenance))
	for _, m := range payload.Entities.Maintenance {
		validMaintenanceIDs[m.ID] = struct{}{}
	}

	validRepairIDs := make(map[uint]struct{}, len(payload.Entities.Repairs))
	for _, r := range payload.Entities.Repairs {
		validRepairIDs[r.ID] = struct{}{}
	}

	for i := range payload.Entities.Maintenance {
		m := &payload.Entities.Maintenance[i]
		if m.ApplianceID != nil {
			if _, ok := validApplianceIDs[*m.ApplianceID]; !ok {
				m.ApplianceID = nil
			}
		}
		if m.AttachmentID != nil {
			if _, ok := validSavedFileIDs[*m.AttachmentID]; !ok {
				m.AttachmentID = nil
			}
		}
	}

	for i := range payload.Entities.Repairs {
		r := &payload.Entities.Repairs[i]
		if r.ApplianceID != nil {
			if _, ok := validApplianceIDs[*r.ApplianceID]; !ok {
				r.ApplianceID = nil
			}
		}
		if r.AttachmentID != nil {
			if _, ok := validSavedFileIDs[*r.AttachmentID]; !ok {
				r.AttachmentID = nil
			}
		}
	}

	for i := range payload.Entities.SavedFiles {
		f := &payload.Entities.SavedFiles[i]
		if f.ApplianceID != nil {
			if _, ok := validApplianceIDs[*f.ApplianceID]; !ok {
				f.ApplianceID = nil
			}
		}
		if f.MaintenanceID != nil {
			if _, ok := validMaintenanceIDs[*f.MaintenanceID]; !ok {
				f.MaintenanceID = nil
			}
		}
		if f.RepairID != nil {
			if _, ok := validRepairIDs[*f.RepairID]; !ok {
				f.RepairID = nil
			}
		}
	}

	for i := range payload.Entities.Notes {
		n := &payload.Entities.Notes[i]
		if n.ApplianceID != nil {
			if _, ok := validApplianceIDs[*n.ApplianceID]; !ok {
				n.ApplianceID = nil
			}
		}
	}

	for i := range payload.Entities.Tasks {
		t := &payload.Entities.Tasks[i]
		if t.ApplianceID != nil {
			if _, ok := validApplianceIDs[*t.ApplianceID]; !ok {
				t.ApplianceID = nil
			}
		}
	}
}
