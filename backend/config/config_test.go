package config

import (
	"os"
	"testing"

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
	// Save original env and restore after test
	originalProxyPort := os.Getenv("PROXY_PORT")
	originalOtelAddr := os.Getenv("OTEL_COLLECTOR_ADDR")
	defer func() {
		os.Setenv("PROXY_PORT", originalProxyPort)
		os.Setenv("OTEL_COLLECTOR_ADDR", originalOtelAddr)
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("PROXY_PORT")
		os.Unsetenv("OTEL_COLLECTOR_ADDR")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, 8080, cfg.ProxyPort)
		assert.Equal(t, "localhost:4317", cfg.Telemetry.CollectorAddr)
	})

	t.Run("override via env", func(t *testing.T) {
		os.Setenv("PROXY_PORT", "9090")
		os.Setenv("OTEL_COLLECTOR_ADDR", "otel:4317")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, 9090, cfg.ProxyPort)
		assert.Equal(t, "otel:4317", cfg.Telemetry.CollectorAddr)
	})
}
