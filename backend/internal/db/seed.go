package db

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
)

// SeedDefaultTenant creates a default tenant if none exist and seeds upstreams. Idempotent.
func SeedDefaultTenant(database *gorm.DB, upstreams []config.UpstreamConfig) error {
	var tenant Tenant
	err := database.Where("name = ?", "default").First(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		key, hash, _, err := auth.GenerateAPIKey()
		if err != nil {
			return fmt.Errorf("seed: generate api key: %w", err)
		}

		tenant = Tenant{
			Name:        "default",
			Description: "Default tenant (auto-seeded)",
			APIKey:      key,
		}
		if err := database.Create(&tenant).Error; err != nil {
			return fmt.Errorf("seed: create default tenant: %w", err)
		}
		slog.Info("seeded default tenant", "id", tenant.ID)

		// Also create a manageable API Key record so it shows up in the UI and works for auth
		apiKey := APIKey{
			Name:     "Default Key",
			KeyHash:  hash,
			Prefix:   auth.KeyPrefix,
			TenantID: tenant.ID,
			IsActive: true,
		}
		if err := database.Create(&apiKey).Error; err != nil {
			slog.Error("failed to seed default api key", "err", err)
		}
	} else if err != nil {
		return fmt.Errorf("seed: check default tenant: %w", err)
	}

	// Seed upstreams from environment for the default tenant
	for _, u := range upstreams {
		var existing Upstream
		err := database.Where("tenant_id = ? AND key_id = ?", tenant.ID, u.KeyID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			up := Upstream{
				KeyID:         u.KeyID,
				ProviderModel: u.Model,
				BaseURL:       u.BaseURL,
				APIKey:        u.APIKey,
				TenantID:      tenant.ID,
			}
			if err := database.Create(&up).Error; err != nil {
				slog.Error("failed to seed upstream", "key_id", u.KeyID, "err", err)
			} else {
				slog.Info("seeded upstream", "key_id", u.KeyID, "tenant", tenant.Name)
			}
		}
	}

	return nil
}
