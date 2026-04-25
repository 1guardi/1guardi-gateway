package config

import (
	"os"
	"strconv"
)

type Config struct {
	ProxyPort int
	AdminPort int
	Telemetry TelemetryConfig
	Redis     RedisConfig
	Postgres  PostgresConfig
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
			DSN: env("POSTGRES_DSN", "postgres://gateway:gateway@localhost:5432/gateway?sslmode=disable"),
		},
	}, nil
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
