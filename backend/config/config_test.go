package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadUpstreams(t *testing.T) {
	keys := []string{
		"UPSTREAM_0_KEY_ID", "UPSTREAM_0_MODEL", "UPSTREAM_0_BASE_URL", "UPSTREAM_0_API_KEY",
		"UPSTREAM_1_KEY_ID",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}

	t.Run("no upstreams configured", func(t *testing.T) {
		ups := loadUpstreams()
		assert.Empty(t, ups)
	})

	t.Run("single upstream", func(t *testing.T) {
		t.Setenv("UPSTREAM_0_KEY_ID", "openai-primary")
		t.Setenv("UPSTREAM_0_MODEL", "gpt-4o")
		t.Setenv("UPSTREAM_0_BASE_URL", "https://api.openai.com")
		t.Setenv("UPSTREAM_0_API_KEY", "sk-test")

		ups := loadUpstreams()
		assert.Len(t, ups, 1)
		assert.Equal(t, "openai-primary", ups[0].KeyID)
		assert.Equal(t, "gpt-4o", ups[0].Model)
		assert.Equal(t, "https://api.openai.com", ups[0].BaseURL)
		assert.Equal(t, "sk-test", ups[0].APIKey)
	})

	t.Run("stops at gap", func(t *testing.T) {
		t.Setenv("UPSTREAM_0_KEY_ID", "first")
		t.Setenv("UPSTREAM_0_MODEL", "gpt-4o")
		// UPSTREAM_1_KEY_ID not set → stops at 1

		ups := loadUpstreams()
		assert.Len(t, ups, 1)
	})

	t.Run("base_url defaults to openai", func(t *testing.T) {
		t.Setenv("UPSTREAM_0_KEY_ID", "k")
		t.Setenv("UPSTREAM_0_MODEL", "gpt-4o")
		// BASE_URL not set

		ups := loadUpstreams()
		assert.Equal(t, "https://api.openai.com", ups[0].BaseURL)
	})
}

func TestLoad(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		// Clear env vars to ensure defaults are used
		vars := []string{"ENV", "PROXY_PORT", "ADMIN_PORT", "OTEL_COLLECTOR_ADDR", "OTEL_SERVICE_NAME", "API_KEY_CACHE_TTL_SEC", "REDIS_ADDR", "POSTGRES_DSN"}
		for _, v := range vars {
			t.Setenv(v, "")
		}

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "prod", cfg.Env)
		assert.Equal(t, 8080, cfg.ProxyPort)
		assert.Equal(t, 8081, cfg.AdminPort)
		assert.Equal(t, "localhost:4317", cfg.Telemetry.CollectorAddr)
		assert.Equal(t, "ai-gateway", cfg.Telemetry.ServiceName)
		assert.Equal(t, 300*time.Second, cfg.Auth.CacheTTL)
		assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
		assert.Equal(t, "postgres://gateway:gateway@localhost:6432/gateway?sslmode=disable", cfg.Postgres.DSN)
	})

	t.Run("override via env", func(t *testing.T) {
		t.Setenv("ENV", "test")
		t.Setenv("PROXY_PORT", "9090")
		t.Setenv("ADMIN_PORT", "9091")
		t.Setenv("OTEL_COLLECTOR_ADDR", "otel:4317")
		t.Setenv("OTEL_SERVICE_NAME", "test-gateway")
		t.Setenv("API_KEY_CACHE_TTL_SEC", "60")
		t.Setenv("REDIS_ADDR", "redis:6379")
		t.Setenv("POSTGRES_DSN", "postgres://user:pass@host:5432/db")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.Env)
		assert.Equal(t, 9090, cfg.ProxyPort)
		assert.Equal(t, 9091, cfg.AdminPort)
		assert.Equal(t, "otel:4317", cfg.Telemetry.CollectorAddr)
		assert.Equal(t, "test-gateway", cfg.Telemetry.ServiceName)
		assert.Equal(t, 60*time.Second, cfg.Auth.CacheTTL)
		assert.Equal(t, "redis:6379", cfg.Redis.Addr)
		assert.Equal(t, "postgres://user:pass@host:5432/db", cfg.Postgres.DSN)
	})
}
