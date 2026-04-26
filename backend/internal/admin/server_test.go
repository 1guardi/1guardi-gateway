package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(database)
	require.NoError(t, err)

	return database
}

func TestTenantHandlers(t *testing.T) {
	database := setupTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(cfg, database, nil)

	// 1. Create a tenant
	newTenant := db.Tenant{
		Name:   "Test Tenant",
		APIKey: "test-key-123",
	}
	body, _ := json.Marshal(newTenant)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var created db.Tenant
	err := json.Unmarshal(rr.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, newTenant.Name, created.Name)
	assert.Equal(t, newTenant.APIKey, created.APIKey)
	assert.NotZero(t, created.ID)

	// 2. List tenants
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var tenants []db.Tenant
	err = json.Unmarshal(rr.Body.Bytes(), &tenants)
	require.NoError(t, err)
	assert.Len(t, tenants, 1)
	assert.Equal(t, created.Name, tenants[0].Name)
}

func TestHandleReady(t *testing.T) {
	database := setupTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(cfg, database, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp healthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ready", resp.Status)
}

func TestHandleListEndpoints_NilRouter(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(&config.Config{}, database, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/router/endpoints", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var endpoints []llmrouter.EndpointStatus
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &endpoints))
	assert.Empty(t, endpoints)
}

func TestHandleListEndpoints_WithRouter(t *testing.T) {
	database := setupTestDB(t)
	r := llmrouter.New([]config.UpstreamConfig{
		{KeyID: "openai-primary", Model: "gpt-4o", BaseURL: "https://api.openai.com", APIKey: "sk-test"},
	})
	handler := NewRouter(&config.Config{}, database, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/router/endpoints", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var endpoints []llmrouter.EndpointStatus
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &endpoints))
	require.Len(t, endpoints, 1)
	assert.Equal(t, "openai-primary", endpoints[0].ID)
	assert.Equal(t, "gpt-4o", endpoints[0].Model)
	assert.Equal(t, "CLOSED", endpoints[0].CircuitState)
}
