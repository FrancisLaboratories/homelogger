package models

import (
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	ID          uint    `json:"id" gorm:"primaryKey"`
	Title       string  `json:"title" gorm:"not null;default:''"`
	Body        string  `json:"body" gorm:"not null;default:''"`
	ApplianceID *uint   `json:"applianceID" gorm:"default:null"`
	SpaceType   *string `json:"spaceType" gorm:"default:null"`
}
