package models

import (
	"gorm.io/gorm"
)

// Task represents a planned maintenance or repair item with optional scheduling and recurrence.
type Task struct {
	gorm.Model
	ID                 uint     `json:"id" gorm:"primaryKey"`
	Label              string   `json:"label" gorm:"not null;default:''"`
	Notes              string   `json:"notes" gorm:"default:''"`
	Checked            bool     `json:"checked" gorm:"default:false;not null"`
	Priority           string   `json:"priority" gorm:"default:''"`
	DueDate            *string  `json:"dueDate" gorm:"default:null"`
	EstimatedCost      *float64 `json:"estimatedCost" gorm:"default:null"`
	IsRecurring        bool     `json:"isRecurring" gorm:"default:false;not null"`
	RecurrenceInterval int      `json:"recurrenceInterval" gorm:"default:0"`
	RecurrenceUnit     string   `json:"recurrenceUnit" gorm:"default:''"`
	RecurrenceMode     string   `json:"recurrenceMode" gorm:"default:''"`
	LastCompletedAt    *string  `json:"lastCompletedAt" gorm:"default:null"`
	UserID             string   `json:"userid" gorm:"not null;default:''"`
	ApplianceID        *uint    `json:"applianceId" gorm:"default:null"`
	SpaceType          *string  `json:"spaceType" gorm:"default:null"`
}
