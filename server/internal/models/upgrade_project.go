package models

import "gorm.io/gorm"

type UpgradeProject struct {
	gorm.Model
	ID            uint           `json:"id" gorm:"primaryKey"`
	Title         string         `json:"title" gorm:"not null"`
	Description   string         `json:"description" gorm:"default:''"`
	Status        string         `json:"status" gorm:"default:'planned'"`
	Priority      string         `json:"priority" gorm:"default:''"`
	TargetDate    string         `json:"targetDate" gorm:"default:''"`
	EstimatedCost float64        `json:"estimatedCost" gorm:"default:0.0"`
	Notes         string         `json:"notes" gorm:"default:''"`
	CategoryID    *uint          `json:"categoryId" gorm:"default:null"`
	Category      BudgetCategory `gorm:"foreignKey:CategoryID;references:ID"`
}
