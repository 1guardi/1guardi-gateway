package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env        string
	ProxyPort  int
	AdminPort  int
	Telemetry  TelemetryConfig
	Auth       APIKeyConfig
	Admin      AdminConfig
	Redis      RedisConfig
	Postgres   PostgresConfig
	ClickHouse ClickHouseConfig
	Upstreams  []UpstreamConfig
	MLRunner   MLRunnerConfig
	OIDC       OIDCConfig
}

// OIDCConfig holds OIDC provider settings. A provider is enabled only when
// both ClientID and ClientSecret are set.
type OIDCConfig struct {
	RedirectBaseURL   string // e.g. "http://localhost:8081" — callback = RedirectBaseURL + /api/v1/auth/oidc/{provider}/callback
	FrontendURL       string // where to bounce the browser after callback w/ token in fragment
	Google            OIDCProviderConfig
	Microsoft         OIDCProviderConfig
	MicrosoftTenantID string // Azure AD tenant ID; "common" allows multi-tenant + personal accounts
}

type OIDCProviderConfig struct {
	ClientID     string
	ClientSecret string
}

func (p OIDCProviderConfig) Enabled() bool {
	return p.ClientID != "" && p.ClientSecret != ""
}

// MLRunnerConfig holds settings for the Python ML inference sidecar.
type MLRunnerConfig struct {
	BaseURL   string        // MLRUNNER_BASE_URL
	Threshold float64       // MLRUNNER_THRESHOLD — min score to flag as injection
	TimeoutMS int           // MLRUNNER_TIMEOUT_MS
	CacheTTL  time.Duration // MLRUNNER_CACHE_TTL_SEC
	Enabled   bool          // MLRUNNER_ENABLED
}

type ClickHouseConfig struct {
	Addr     string // native protocol address, e.g. "localhost:9000"
	User     string
	Password string
	Database string
}

type AdminConfig struct {
	Email       string
	Password    string
	JWTSecret   string
	JWTTTLHours int
}

// UpstreamConfig describes one LLM endpoint the gateway can route to.
// Multiple entries with the same Model enable fallback routing.
type UpstreamConfig struct {
	KeyID    string // unique label, e.g. "openai-primary"
	Provider string // e.g. "openai", "anthropic", "gemini"
	Model    string // model name, e.g. "gpt-4o"
	BaseURL  string // e.g. "https://api.openai.com"
	APIKey   string
	TenantID uint // 0 means default/global if applicable, but usually tied to a tenant
}

type TelemetryConfig struct {
	CollectorAddr string
	ServiceName   string
}

type RedisConfig struct {
	Addr string
}

type APIKeyConfig struct {
	CacheTTL time.Duration
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
		Auth: APIKeyConfig{
			CacheTTL: time.Duration(intEnv("API_KEY_CACHE_TTL_SEC", 300)) * time.Second,
		},
		Admin: AdminConfig{
			Email:       env("ADMIN_EMAIL", "admin@example.com"),
			Password:    env("ADMIN_PASSWORD", ""),
			JWTSecret:   loadJWTSecret(),
			JWTTTLHours: intEnv("JWT_TTL_HOURS", 24),
		},
		Redis: RedisConfig{
			Addr: env("REDIS_ADDR", "localhost:6379"),
		},
		Postgres: PostgresConfig{
			DSN: env("POSTGRES_DSN", "postgres://gateway:gateway@localhost:6432/gateway?sslmode=disable"),
		},
		ClickHouse: ClickHouseConfig{
			Addr:     env("CLICKHOUSE_ADDR", "localhost:9001"),
			User:     env("CLICKHOUSE_USER", "default"),
			Password: env("CLICKHOUSE_PASSWORD", "otel"),
			Database: env("CLICKHOUSE_DB", "otel"),
		},
		Upstreams: loadUpstreams(),
		OIDC: OIDCConfig{
			RedirectBaseURL:   env("OIDC_REDIRECT_BASE_URL", "http://localhost:8081"),
			FrontendURL:       env("OIDC_FRONTEND_URL", "http://localhost:5173"),
			MicrosoftTenantID: env("OIDC_MICROSOFT_TENANT", "common"),
			Google: OIDCProviderConfig{
				ClientID:     os.Getenv("OIDC_GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("OIDC_GOOGLE_CLIENT_SECRET"),
			},
			Microsoft: OIDCProviderConfig{
				ClientID:     os.Getenv("OIDC_MICROSOFT_CLIENT_ID"),
				ClientSecret: os.Getenv("OIDC_MICROSOFT_CLIENT_SECRET"),
			},
		},
		MLRunner: MLRunnerConfig{
			BaseURL:   env("MLRUNNER_BASE_URL", "http://localhost:8082"),
			Threshold: floatEnv("MLRUNNER_THRESHOLD", 0.85),
			TimeoutMS: intEnv("MLRUNNER_TIMEOUT_MS", 200),
			CacheTTL:  time.Duration(intEnv("MLRUNNER_CACHE_TTL_SEC", 86400)) * time.Second,
			Enabled:   boolEnv("MLRUNNER_ENABLED", true),
		},
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
			KeyID:    keyID,
			Provider: env(prefix+"PROVIDER", "openai"),
			Model:    os.Getenv(prefix + "MODEL"),
			BaseURL:  env(prefix+"BASE_URL", "https://api.openai.com"),
			APIKey:   os.Getenv(prefix + "API_KEY"),
		})
	}
	return upstreams
}

func loadJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	slog.Warn("JWT_SECRET not set — generating random secret; tokens invalidated on restart")
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate JWT secret: " + err.Error())
	}
	return hex.EncodeToString(b)
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

func floatEnv(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func boolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
