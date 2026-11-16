package database

import (
	"os"

	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

const improvementIDPredicate = "improvement_id = ?"

// FileInfoResponse represents the response structure for file information
type FileInfoResponse struct {
	ID           uint   `json:"id"`
	OriginalName string `json:"originalName"`
	UserID       string `json:"userID"`
}

// GetFileInfo gets a file by ID and returns only id, originalName, and userID
func GetFileInfo(db *gorm.DB, id uint) (*FileInfoResponse, error) {
	var file models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where("id = ?", id).First(&file)
	if result.Error != nil {
		return nil, result.Error
	}

	return &FileInfoResponse{
		ID:           file.ID,
		OriginalName: file.OriginalName,
		UserID:       file.UserID,
	}, nil
}

// GetFilePath retrieves the file path of a file by its ID
func GetFilePath(db *gorm.DB, id uint) (string, error) {
	var file models.SavedFile
	result := db.Select("path").Where("id = ?", id).First(&file)
	if result.Error != nil {
		return "", result.Error
	}
	return file.Path, nil
}

// UploadFile uploads a new file
func UploadFile(db *gorm.DB, file *models.SavedFile) (*models.SavedFile, error) {
	// Ensure the ID is not manually set
	file.ID = 0
	result := db.Create(file)
	if result.Error != nil {
		return nil, result.Error
	}

	return file, nil
}

// UpdateFilePath updates the file path of an existing file
func UpdateFilePath(db *gorm.DB, file *models.SavedFile) (*models.SavedFile, error) {
	result := db.Model(&models.SavedFile{}).Where("id = ?", file.ID).Update("path", file.Path)
	if result.Error != nil {
		return nil, result.Error
	}

	return file, nil
}

// AttachFileToMaintenance sets the maintenance_id for a saved file
func AttachFileToMaintenance(db *gorm.DB, fileID uint, maintenanceID uint) error {
	result := db.Model(&models.SavedFile{}).Where("id = ?", fileID).Update("maintenance_id", maintenanceID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// AttachFileToRepair sets the repair_id for a saved file
func AttachFileToRepair(db *gorm.DB, fileID uint, repairID uint) error {
	result := db.Model(&models.SavedFile{}).Where("id = ?", fileID).Update("repair_id", repairID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func AttachFileToImprovement(db *gorm.DB, fileID uint, improvementID uint) error {
	result := db.Model(&models.SavedFile{}).Where("id = ?", fileID).Update("improvement_id", improvementID)
	return result.Error
}

// DeleteFile deletes the file record by ID
func DeleteFile(db *gorm.DB, id uint) error {
	result := db.Unscoped().Where("id = ?", id).Delete(&models.SavedFile{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetFilesByMaintenance returns file info for files attached to a maintenance record
func GetFilesByMaintenance(db *gorm.DB, maintenanceID uint) ([]FileInfoResponse, error) {
	var files []models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where("maintenance_id = ?", maintenanceID).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := make([]FileInfoResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, FileInfoResponse{ID: f.ID, OriginalName: f.OriginalName, UserID: f.UserID})
	}
	return resp, nil
}

// GetFilesByRepair returns file info for files attached to a repair record
func GetFilesByRepair(db *gorm.DB, repairID uint) ([]FileInfoResponse, error) {
	var files []models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where("repair_id = ?", repairID).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := make([]FileInfoResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, FileInfoResponse{ID: f.ID, OriginalName: f.OriginalName, UserID: f.UserID})
	}
	return resp, nil
}

func GetFilesByImprovement(db *gorm.DB, improvementID uint) ([]FileInfoResponse, error) {
	var files []models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where(improvementIDPredicate, improvementID).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}
	resp := make([]FileInfoResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, FileInfoResponse{ID: f.ID, OriginalName: f.OriginalName, UserID: f.UserID})
	}
	return resp, nil
}

// DeleteFilesByMaintenance removes files on disk and deletes their DB rows for a maintenance record
func DeleteFilesByMaintenance(db *gorm.DB, maintenanceID uint) error {
	var files []models.SavedFile
	result := db.Select("id", "path").Where("maintenance_id = ?", maintenanceID).Find(&files)
	if result.Error != nil {
		return result.Error
	}

	for _, f := range files {
		if f.Path != "" {
			_ = os.Remove(f.Path)
		}
	}

	// Delete DB rows (hard delete)
	if err := db.Unscoped().Where("maintenance_id = ?", maintenanceID).Delete(&models.SavedFile{}).Error; err != nil {
		return err
	}
	return nil
}

// DeleteFilesByRepair removes files on disk and deletes their DB rows for a repair record
func DeleteFilesByRepair(db *gorm.DB, repairID uint) error {
	var files []models.SavedFile
	result := db.Select("id", "path").Where("repair_id = ?", repairID).Find(&files)
	if result.Error != nil {
		return result.Error
	}

	for _, f := range files {
		if f.Path != "" {
			_ = os.Remove(f.Path)
		}
	}

	// Delete DB rows (hard delete)
	if err := db.Unscoped().Where("repair_id = ?", repairID).Delete(&models.SavedFile{}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteFilesByImprovement(db *gorm.DB, improvementID uint) error {
	var files []models.SavedFile
	result := db.Select("id", "path").Where(improvementIDPredicate, improvementID).Find(&files)
	if result.Error != nil {
		return result.Error
	}
	for _, f := range files {
		if f.Path != "" {
			_ = os.Remove(f.Path)
		}
	}
	return db.Unscoped().Where(improvementIDPredicate, improvementID).Delete(&models.SavedFile{}).Error
}

// DeleteFilesByAppliance removes files on disk and deletes their DB rows for an appliance
func DeleteFilesByAppliance(db *gorm.DB, applianceID uint) error {
	var files []models.SavedFile
	result := db.Select("id", "path").Where("appliance_id = ?", applianceID).Find(&files)
	if result.Error != nil {
		return result.Error
	}

	for _, f := range files {
		if f.Path != "" {
			_ = os.Remove(f.Path)
		}
	}

	// Delete DB rows (hard delete)
	if err := db.Unscoped().Where("appliance_id = ?", applianceID).Delete(&models.SavedFile{}).Error; err != nil {
		return err
	}
	return nil
}

// GetFilesByAppliance returns file info for files attached to an appliance
func GetFilesByAppliance(db *gorm.DB, applianceID uint) ([]FileInfoResponse, error) {
	var files []models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where("appliance_id = ?", applianceID).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := make([]FileInfoResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, FileInfoResponse{ID: f.ID, OriginalName: f.OriginalName, UserID: f.UserID})
	}
	return resp, nil
}

// AttachFileToAppliance sets the appliance_id for a saved file
func AttachFileToAppliance(db *gorm.DB, fileID uint, applianceID uint) error {
	result := db.Model(&models.SavedFile{}).Where("id = ?", fileID).Update("appliance_id", applianceID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// AttachFileToSpace sets the space_type for a saved file
func AttachFileToSpace(db *gorm.DB, fileID uint, spaceType string) error {
	result := db.Model(&models.SavedFile{}).Where("id = ?", fileID).Update("space_type", spaceType)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetFilesBySpace returns file info for files attached to a space type
func GetFilesBySpace(db *gorm.DB, spaceType string) ([]FileInfoResponse, error) {
	var files []models.SavedFile
	result := db.Select("id", "original_name", "user_id").Where("space_type = ?", spaceType).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := make([]FileInfoResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, FileInfoResponse{ID: f.ID, OriginalName: f.OriginalName, UserID: f.UserID})
	}
	return resp, nil
}

// DeleteFilesBySpace removes files on disk and deletes their DB rows for a space type
func DeleteFilesBySpace(db *gorm.DB, spaceType string) error {
	var files []models.SavedFile
	result := db.Select("id", "path").Where("space_type = ?", spaceType).Find(&files)
	if result.Error != nil {
		return result.Error
	}

	for _, f := range files {
		if f.Path != "" {
			_ = os.Remove(f.Path)
		}
	}

	// Delete DB rows (hard delete)
	if err := db.Unscoped().Where("space_type = ?", spaceType).Delete(&models.SavedFile{}).Error; err != nil {
		return err
	}
	return nil
}
