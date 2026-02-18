package models

import "gorm.io/gorm"

type BudgetCategory struct {
	gorm.Model
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	AssetGroup  string `json:"assetGroup" gorm:"default:''"`
	Description string `json:"description" gorm:"default:''"`
	Color       string `json:"color" gorm:"default:''"`
}
