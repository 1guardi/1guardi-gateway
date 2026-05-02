package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

func setupSeedTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(database))
	return database
}

func TestSeedDefaultTenant_CreatesWhenEmpty(t *testing.T) {
	database := setupSeedTestDB(t)

	err := SeedDefaultTenant(database, nil)
	require.NoError(t, err)

	var tenants []Tenant
	database.Find(&tenants)
	require.Len(t, tenants, 1)
	assert.Equal(t, "default", tenants[0].Name)
	assert.NotEmpty(t, tenants[0].APIKey)
}

func TestSeedDefaultTenant_IdempotentWhenExists(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedDefaultTenant(database, nil))
	require.NoError(t, SeedDefaultTenant(database, nil))

	var count int64
	database.Model(&Tenant{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestSeedDefaultTenant_WithUpstreams(t *testing.T) {
	database := setupSeedTestDB(t)

	upstreams := []config.UpstreamConfig{
		{KeyID: "test-ups-1", Model: "gpt-4o", BaseURL: "http://test", APIKey: "sk-test"},
	}

	err := SeedDefaultTenant(database, upstreams)
	require.NoError(t, err)

	var ups []Upstream
	database.Find(&ups)
	require.Len(t, ups, 1)
	assert.Equal(t, "test-ups-1", ups[0].KeyID)
	assert.Equal(t, "gpt-4o", ups[0].Models)

	// Idempotent test
	err = SeedDefaultTenant(database, upstreams)
	require.NoError(t, err)

	database.Find(&ups)
	require.Len(t, ups, 1, "Should not duplicate upstreams")
}

func TestSeedSuperAdmin_Creates(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", "secret"))

	var user User
	require.NoError(t, database.Where("email = ?", "admin@example.com").First(&user).Error)
	assert.Equal(t, "Super Admin", user.Name)
	assert.Equal(t, "admin@example.com", user.Email)
	assert.True(t, user.IsSuperAdmin)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret")))
}

func TestSeedSuperAdmin_Idempotent(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", "secret"))
	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", "secret"))

	var count int64
	database.Model(&User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestSeedSuperAdmin_UpdatesPassword(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", "old-pass"))
	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", "new-pass"))

	var user User
	require.NoError(t, database.Where("email = ?", "admin@example.com").First(&user).Error)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("new-pass")))
}

func TestSeedSuperAdmin_EmptyPassword(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedSuperAdmin(database, "admin@example.com", ""))

	var count int64
	database.Model(&User{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestSeedRBAC(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedRBAC(database))

	var role Role
	require.NoError(t, database.Where("name = ?", "tenantAdmin").Preload("Permissions").First(&role).Error)
	assert.NotEmpty(t, role.Permissions)

	var userRole Role
	require.NoError(t, database.Where("name = ?", "user").Preload("Permissions").First(&userRole).Error)
	assert.NotEmpty(t, userRole.Permissions)

	// Check that user role has fewer permissions than admin (it's read-only)
	assert.True(t, len(userRole.Permissions) < len(role.Permissions))

	// Test idempotency
	require.NoError(t, SeedRBAC(database))
}
