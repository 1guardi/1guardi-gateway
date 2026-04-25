package db

import (
	"gorm.io/gorm"
)

// Tenant represents an LLM consumer or project.
type Tenant struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	APIKey      string `gorm:"uniqueIndex;not null"` // Internal key for this tenant to use the gateway
}

// AutoMigrate runs schema migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
	)
}
