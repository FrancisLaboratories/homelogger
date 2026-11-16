package models

import "gorm.io/gorm"

type Improvement struct {
	gorm.Model
	ID            uint      `json:"id" gorm:"primaryKey"`
	Description   string    `json:"description" gorm:"not null;default:''"`
	Date          string    `json:"date" gorm:"not null;default:''"`
	Cost          float64   `json:"cost" gorm:"not null;default:0"`
	Notes         string    `json:"notes" gorm:"not null;default:''"`
	SpaceType     string    `json:"spaceType" gorm:"not null;default:''"`
	ReferenceType string    `json:"referenceType" gorm:"not null;default:''"`
	ApplianceID   *uint     `json:"applianceId" gorm:"default:null"`
	Appliance     Appliance `gorm:"foreignKey:ApplianceID;references:ID"`
}
