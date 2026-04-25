package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env       string
	ProxyPort int
	AdminPort int
	Telemetry TelemetryConfig
	Redis     RedisConfig
	Postgres  PostgresConfig
	Upstreams []UpstreamConfig
}

// UpstreamConfig describes one LLM endpoint the gateway can route to.
// Multiple entries with the same Model enable fallback routing.
type UpstreamConfig struct {
	KeyID   string // unique label, e.g. "openai-primary"
	Model   string // model name, e.g. "gpt-4o"
	BaseURL string // e.g. "https://api.openai.com"
	APIKey  string
}

type TelemetryConfig struct {
	CollectorAddr string
	ServiceName   string
}

type RedisConfig struct {
	Addr string
}

type PostgresConfig struct {
	DSN string
}

func Load() (*Config, error) {
	return &Config{
		Env:       env("ENV", "prod"),
		ProxyPort: intEnv("PROXY_PORT", 8080),
		AdminPort: intEnv("ADMIN_PORT", 8081),
		Telemetry: TelemetryConfig{
			CollectorAddr: env("OTEL_COLLECTOR_ADDR", "localhost:4317"),
			ServiceName:   env("OTEL_SERVICE_NAME", "ai-gateway"),
		},
		Redis: RedisConfig{
			Addr: env("REDIS_ADDR", "localhost:6379"),
		},
		Postgres: PostgresConfig{
			DSN: env("POSTGRES_DSN", "postgres://gateway:gateway@localhost:6432/gateway?sslmode=disable"),
		},
		Upstreams: loadUpstreams(),
	}, nil
}

// loadUpstreams reads UPSTREAM_N_* env vars (N=0..9) until a gap is found.
func loadUpstreams() []UpstreamConfig {
	var upstreams []UpstreamConfig
	for i := 0; i < 10; i++ {
		prefix := fmt.Sprintf("UPSTREAM_%d_", i)
		keyID := os.Getenv(prefix + "KEY_ID")
		if keyID == "" {
			break
		}
		upstreams = append(upstreams, UpstreamConfig{
			KeyID:   keyID,
			Model:   os.Getenv(prefix + "MODEL"),
			BaseURL: env(prefix+"BASE_URL", "https://api.openai.com"),
			APIKey:  os.Getenv(prefix + "API_KEY"),
		})
	}
	return upstreams
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
