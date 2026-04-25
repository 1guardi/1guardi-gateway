package db

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

// Setup initializes the database connection and performs migrations.
func Setup(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Postgres.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("connected to database", "dsn", cfg.Postgres.DSN)

	if cfg.Env == "local" {
		// Run migrations
		if err := AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return db, nil
}
