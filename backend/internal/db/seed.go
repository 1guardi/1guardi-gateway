package db

import (
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
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
				KeyID:    u.KeyID,
				Provider: u.Provider,
				Models:   u.Model, // u.Model from config is a single model, store it as initial model
				BaseURL:  u.BaseURL,
				APIKey:   u.APIKey,
				TenantID: tenant.ID,
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

// SeedSuperAdmin creates or updates the super admin user from env config. Idempotent.
func SeedSuperAdmin(database *gorm.DB, email, password string) error {
	if password == "" {
		slog.Warn("ADMIN_PASSWORD not set — admin login disabled")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("seed: hash admin password: %w", err)
	}

	var existing User
	err = database.Where("email = ?", email).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		user := User{
			Name:         "Super Admin",
			Email:        email,
			PasswordHash: string(hash),
			IsSuperAdmin: true,
		}
		if err := database.Create(&user).Error; err != nil {
			return fmt.Errorf("seed: create admin user: %w", err)
		}
		slog.Info("seeded super admin user", "email", email)
		return nil
	} else if err != nil {
		return fmt.Errorf("seed: check admin user: %w", err)
	}

	// Ensure IsSuperAdmin is set
	if !existing.IsSuperAdmin {
		if err := database.Model(&existing).Update("is_super_admin", true).Error; err != nil {
			return fmt.Errorf("seed: update admin user is_super_admin: %w", err)
		}
	}

	// Update password hash if the plaintext password has changed
	if bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(password)) != nil {
		if err := database.Model(&existing).Update("password_hash", string(hash)).Error; err != nil {
			return fmt.Errorf("seed: update admin password: %w", err)
		}
		slog.Info("updated super admin user password", "email", email)
	}

	return nil
}

// SeedRBAC creates default roles and permissions. Idempotent.
func SeedRBAC(database *gorm.DB) error {
	permissions := []string{
		"tenant.read",
		"tenant.update",
		"agents.read",
		"agents.manage",
		"keys.read",
		"keys.manage",
		"upstreams.read",
		"upstreams.manage",
		"members.read",
		"members.manage",
	}

	// Seed Permissions
	for _, pName := range permissions {
		var p Permission
		if err := database.Where("name = ?", pName).FirstOrCreate(&p, Permission{Name: pName}).Error; err != nil {
			return fmt.Errorf("seed: create permission %s: %w", pName, err)
		}
	}

	var allPerms []Permission
	database.Find(&allPerms)

	var readPerms []Permission
	database.Where("name LIKE ?", "%.read").Find(&readPerms)

	// Seed Roles
	var tenantAdmin Role
	if err := database.Where("name = ?", "tenantAdmin").FirstOrCreate(&tenantAdmin, Role{Name: "tenantAdmin", Description: "Full access to the tenant"}).Error; err != nil {
		return fmt.Errorf("seed: create role tenantAdmin: %w", err)
	}
	// Replace permissions
	database.Model(&tenantAdmin).Association("Permissions").Replace(allPerms)

	var user Role
	if err := database.Where("name = ?", "user").FirstOrCreate(&user, Role{Name: "user", Description: "Read-only access to the tenant"}).Error; err != nil {
		return fmt.Errorf("seed: create role user: %w", err)
	}
	database.Model(&user).Association("Permissions").Replace(readPerms)

	slog.Info("seeded RBAC roles and permissions")
	return nil
}
