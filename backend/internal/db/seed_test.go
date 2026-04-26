package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSeedTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(database))
	return database
}

func TestSeedDefaultTenant_CreatesWhenEmpty(t *testing.T) {
	database := setupSeedTestDB(t)

	err := SeedDefaultTenant(database)
	require.NoError(t, err)

	var tenants []Tenant
	database.Find(&tenants)
	require.Len(t, tenants, 1)
	assert.Equal(t, "default", tenants[0].Name)
	assert.NotEmpty(t, tenants[0].APIKey)
}

func TestSeedDefaultTenant_IdempotentWhenExists(t *testing.T) {
	database := setupSeedTestDB(t)

	require.NoError(t, SeedDefaultTenant(database))
	require.NoError(t, SeedDefaultTenant(database))

	var count int64
	database.Model(&Tenant{}).Count(&count)
	assert.Equal(t, int64(1), count)
}
