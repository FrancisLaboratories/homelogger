package models

import "gorm.io/gorm"

type BudgetScenario struct {
	gorm.Model
	ID             uint    `json:"id" gorm:"primaryKey"`
	Name           string  `json:"name" gorm:"not null"`
	StartDate      string  `json:"startDate" gorm:"default:''"`
	HorizonMonths  int     `json:"horizonMonths" gorm:"default:0"`
	InflationRate  float64 `json:"inflationRate" gorm:"default:0.0"`
	IsActive       bool    `json:"isActive" gorm:"default:false"`
	Notes          string  `json:"notes" gorm:"default:''"`
}
