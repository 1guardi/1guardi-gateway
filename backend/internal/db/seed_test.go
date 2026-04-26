package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "gpt-4o", ups[0].ProviderModel)

	// Idempotent test
	err = SeedDefaultTenant(database, upstreams)
	require.NoError(t, err)

	database.Find(&ups)
	require.Len(t, ups, 1, "Should not duplicate upstreams")
}
