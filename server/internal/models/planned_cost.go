package models

import "gorm.io/gorm"

type PlannedCost struct {
	gorm.Model
	ID          uint           `json:"id" gorm:"primaryKey"`
	ScenarioID  *uint          `json:"scenarioId" gorm:"default:null"`
	Scenario    BudgetScenario `gorm:"foreignKey:ScenarioID;references:ID"`
	CategoryID  *uint          `json:"categoryId" gorm:"default:null"`
	Category    BudgetCategory `gorm:"foreignKey:CategoryID;references:ID"`
	SourceType  string         `json:"sourceType" gorm:"default:''"`
	SourceID    *uint          `json:"sourceId" gorm:"default:null"`
	CostDate    string         `json:"costDate" gorm:"default:''"`
	Amount      float64        `json:"amount" gorm:"default:0.0"`
	Notes       string         `json:"notes" gorm:"default:''"`
}
