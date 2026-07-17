package models

import "time"

// BackupPayload is the root structure for database backups in JSON format.
type BackupPayload struct {
	Version      string    `json:"version"`
	ExportedAt   time.Time `json:"exportedAt"`
	DatabaseType string    `json:"databaseType"` // "sqlite" or "postgresql"
	Entities     Entities  `json:"entities"`
}

// Entities holds all exported database tables.
type Entities struct {
	Appliances   []Appliance   `json:"appliances"`
	Tasks        []Task        `json:"tasks"`
	Maintenance  []Maintenance `json:"maintenance"`
	Repairs      []Repair      `json:"repairs"`
	SavedFiles   []SavedFile   `json:"savedFiles"`
	Notes        []Note        `json:"notes"`
	Todos        []Todo        `json:"todos"`
}

// ImportResult summarizes the results of an import operation.
type ImportResult struct {
	Inserted     int    `json:"inserted"`
	Updated      int    `json:"updated"`
	Skipped      int    `json:"skipped"`
	Errors       int    `json:"errors"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}


