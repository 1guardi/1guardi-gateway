package db

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
)

// SeedDefaultTenant creates a default tenant if none exist. Idempotent.
func SeedDefaultTenant(database *gorm.DB) error {
	var count int64
	if err := database.Model(&Tenant{}).Count(&count).Error; err != nil {
		return fmt.Errorf("seed: count tenants: %w", err)
	}
	if count > 0 {
		return nil
	}

	key, _, err := auth.GenerateAPIKey()
	if err != nil {
		return fmt.Errorf("seed: generate api key: %w", err)
	}

	tenant := &Tenant{
		Name:        "default",
		Description: "Default tenant (auto-seeded)",
		APIKey:      key,
	}
	if err := database.Create(tenant).Error; err != nil {
		return fmt.Errorf("seed: create default tenant: %w", err)
	}

	slog.Info("seeded default tenant", "id", tenant.ID)
	return nil
}
