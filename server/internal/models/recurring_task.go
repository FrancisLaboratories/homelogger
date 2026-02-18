package models

import "gorm.io/gorm"

type RecurringTask struct {
	gorm.Model
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Description    string         `json:"description" gorm:"default:''"`
	IntervalValue  int            `json:"intervalValue" gorm:"default:1"`
	IntervalUnit   string         `json:"intervalUnit" gorm:"default:'month'"`
	NextDueDate    string         `json:"nextDueDate" gorm:"default:''"`
	EstimatedCost  float64        `json:"estimatedCost" gorm:"default:0.0"`
	ReferenceType  string         `json:"referenceType" gorm:"default:''"`
	SpaceType      string         `json:"spaceType" gorm:"default:''"`
	ApplianceID    *uint          `json:"applianceId" gorm:"default:null"`
	Appliance      Appliance      `gorm:"foreignKey:ApplianceID;references:ID"`
	CategoryID     *uint          `json:"categoryId" gorm:"default:null"`
	Category       BudgetCategory `gorm:"foreignKey:CategoryID;references:ID"`
	AutoCreateTodo bool           `json:"autoCreateTodo" gorm:"default:false"`
	Notes          string         `json:"notes" gorm:"default:''"`
}
