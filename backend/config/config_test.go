package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
