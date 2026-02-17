package models

import "gorm.io/gorm"

type Settings struct {
	gorm.Model
	ID                uint   `json:"id" gorm:"primaryKey"`
	Locale            string `json:"locale" gorm:"not null;default:'en-US'"`
	Language          string `json:"language" gorm:"not null;default:'en'"`
	Currency          string `json:"currency" gorm:"not null;default:'USD'"`
	TimeZone          string `json:"timeZone" gorm:"not null;default:'UTC'"`
	MeasurementSystem string `json:"measurementSystem" gorm:"not null;default:'imperial'"`
	WeekStart         int    `json:"weekStart" gorm:"not null;default:0"`
	DateFormat        string `json:"dateFormat" gorm:"not null;default:''"`
	NumberingSystem   string `json:"numberingSystem" gorm:"not null;default:'latn'"`
}
