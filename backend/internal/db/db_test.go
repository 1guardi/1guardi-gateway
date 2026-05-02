package db

import (
	"testing"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetup_SQLite(t *testing.T) {
	// We override the Postgres driver with SQLite for testing if we wanted to test Setup directly,
	// but Setup is hardcoded to Postgres.
	// Let's at least test AutoMigrate which is called by Setup.

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = AutoMigrate(db)
	assert.NoError(t, err)

	// Check if tables exist
	assert.True(t, db.Migrator().HasTable(&Tenant{}))
	assert.True(t, db.Migrator().HasTable(&Agent{}))
	assert.True(t, db.Migrator().HasTable(&APIKey{}))
	assert.True(t, db.Migrator().HasTable(&Upstream{}))
}

func TestSetup_Error(t *testing.T) {
	cfg := config.Config{}
	cfg.Postgres.DSN = "invalid dsn"
	_, err := Setup(cfg)
	assert.Error(t, err)
}
