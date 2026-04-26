package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
)

func TestAuthenticate(t *testing.T) {
	// Setup test DB
	database, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.AutoMigrate(&db.Tenant{}, &db.Agent{}, &db.APIKey{})

	// Create a test tenant and key
	tenant := db.Tenant{Name: "test-tenant"}
	database.Create(&tenant)

	rawKey, hash, _ := auth.GenerateAPIKey()
	apiKey := db.APIKey{
		Name:     "test-key",
		KeyHash:  hash,
		Prefix:   auth.KeyPrefix,
		TenantID: tenant.ID,
		IsActive: true,
	}
	database.Create(&apiKey)

	srv := &Server{db: database}
	middleware := srv.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantCtx(r.Context())
		assert.Equal(t, fmt.Sprintf("%d", tenant.ID), tc.TenantID)
		assert.Equal(t, "thread-789", tc.ThreadID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	req.Header.Set("X-Thread-Id", "thread-789")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticate_ProjectKeyWithHeader(t *testing.T) {
	database, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.AutoMigrate(&db.Tenant{}, &db.Agent{}, &db.APIKey{})

	tenant := db.Tenant{Name: "test-tenant"}
	database.Create(&tenant)

	rawKey, hash, _ := auth.GenerateAPIKey()
	apiKey := db.APIKey{
		KeyHash:  hash,
		TenantID: tenant.ID,
		IsActive: true,
	}
	database.Create(&apiKey)

	srv := &Server{db: database}
	middleware := srv.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantCtx(r.Context())
		assert.Equal(t, "custom-agent-id", tc.AgentID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	req.Header.Set("X-Agent-Id", "custom-agent-id")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticate_ScopedKey(t *testing.T) {
	database, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.AutoMigrate(&db.Tenant{}, &db.Agent{}, &db.APIKey{})

	tenant := db.Tenant{Name: "test-tenant"}
	database.Create(&tenant)

	agent := db.Agent{Name: "support", TenantID: tenant.ID}
	database.Create(&agent)

	rawKey, hash, _ := auth.GenerateAPIKey()
	agentID := agent.ID
	apiKey := db.APIKey{
		KeyHash:  hash,
		TenantID: tenant.ID,
		AgentID:  &agentID,
		IsActive: true,
	}
	database.Create(&apiKey)

	srv := &Server{db: database}

	t.Run("autofill agent id", func(t *testing.T) {
		middleware := srv.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tc := TenantCtx(r.Context())
			assert.Equal(t, fmt.Sprintf("%d", agent.ID), tc.AgentID)
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("mismatching agent id", func(t *testing.T) {
		middleware := srv.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		req.Header.Set("X-Agent-Id", "wrong-agent")
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func TestNewRouter(t *testing.T) {
	database, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	cfg := &config.Config{}

	// The NewRouter should create a fully wired chi router.
	handler := NewRouter(cfg, database, nil, nil)

	// Because of the Authenticate middleware, any request without a valid token
	// should be blocked with 401 Unauthorized.

	endpoints := []string{
		"/v1/chat/completions",
		"/v1/completions",
		"/v1/embeddings",
		"/v1/models",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ep, nil)
			if ep == "/v1/models" {
				req = httptest.NewRequest(http.MethodGet, ep, nil)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Auth should block it before routing
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}
