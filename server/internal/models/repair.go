package models

import (
	"gorm.io/gorm"
)

type Repair struct {
	gorm.Model
	ID            uint      `json:"id" gorm:"primaryKey"`
	Description   string    `json:"description" gorm:"not null" gorm:"default:''"`
	Date          string    `json:"date" gorm:"not null" gorm:"default:''"`
	Cost          float64   `json:"cost" gorm:"not null" gorm:"default:0.0"`
	Notes         string    `json:"notes" gorm:"not null" gorm:"default:''"`
	SpaceType     string    `json:"spaceType" gorm:"not null" gorm:"default:''"`
	ReferenceType string    `json:"referenceType" gorm:"not null" gorm:"default:''"`
	ApplianceID   *uint     `json:"applianceId" gorm:"default:null"`
	Appliance     Appliance `gorm:"foreignKey:ApplianceID;references:ID"` // Foreign key
	AttachmentID  *uint     `json:"attachmentId" gorm:"default:null"`
	Attachment    SavedFile `gorm:"foreignKey:AttachmentID;references:ID"`
}
