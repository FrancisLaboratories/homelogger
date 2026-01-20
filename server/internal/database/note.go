package database

import (
	"github.com/masoncfrancis/homelogger/server/internal/models"
	"gorm.io/gorm"
)

// GetNotes returns notes filtered by optional applianceId and spaceType.
// Pass applianceId=0 and spaceType="" for no filter.
func GetNotes(db *gorm.DB, applianceId uint, spaceType string) ([]models.Note, error) {
	var notes []models.Note
	query := db.Model(&models.Note{})

	if applianceId != 0 {
		query = query.Where("appliance_id = ?", applianceId)
	}
	if spaceType != "" {
		query = query.Where("space_type = ?", spaceType)
	}

	result := query.Find(&notes)
	if result.Error != nil {
		return nil, result.Error
	}
	return notes, nil
}

// AddNote creates and returns a note
func AddNote(db *gorm.DB, title string, body string, applianceId uint, spaceType string) (models.Note, error) {
	note := models.Note{Title: title, Body: body}
	if applianceId != 0 {
		note.ApplianceID = &applianceId
	}
	if spaceType != "" {
		note.SpaceType = &spaceType
	}

	result := db.Create(&note)
	if result.Error != nil {
		return models.Note{}, result.Error
	}
	return note, nil
}

// GetNote returns a single note by id
func GetNote(db *gorm.DB, id uint) (models.Note, error) {
	var note models.Note
	result := db.First(&note, id)
	if result.Error != nil {
		return models.Note{}, result.Error
	}
	return note, nil
}

// UpdateNote updates title/body of a note and returns the updated note
func UpdateNote(db *gorm.DB, id uint, title string, body string) (models.Note, error) {
	var note models.Note
	if err := db.First(&note, id).Error; err != nil {
		return models.Note{}, err
	}
	note.Title = title
	note.Body = body
	if err := db.Save(&note).Error; err != nil {
		return models.Note{}, err
	}
	return note, nil
}

// DeleteNote deletes a note
func DeleteNote(db *gorm.DB, id uint) error {
	result := db.Where("id = ?", id).Delete(&models.Note{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
