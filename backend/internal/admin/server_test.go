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

func TestAgentHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(&config.Config{}, database, nil)

	// Setup Tenant
	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Agent
	agentBody := `{"Name":"agent1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/agents", bytes.NewBufferString(agentBody))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// List Agents
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/agents", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var agents []db.Agent
	json.Unmarshal(rr.Body.Bytes(), &agents)
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent1", agents[0].Name)
}

func TestKeyHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(&config.Config{}, database, nil)

	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Key
	keyBody := `{"name":"test-key"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/keys", bytes.NewBufferString(keyBody))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var created struct {
		ID  uint   `json:"ID"`
		Key string `json:"key"`
	}
	json.Unmarshal(rr.Body.Bytes(), &created)
	assert.NotEmpty(t, created.Key)

	// List Keys
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/keys", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var keys []db.APIKey
	json.Unmarshal(rr.Body.Bytes(), &keys)
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key", keys[0].Name)
	assert.Empty(t, keys[0].KeyHash, "KeyHash should be hidden")

	// Revoke Key
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/1/keys/1", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	var revokedKey db.APIKey
	database.First(&revokedKey, 1)
	assert.False(t, revokedKey.IsActive)
}

func TestUpstreamHandlers(t *testing.T) {
	database := setupTestDB(t)
	r := llmrouter.New([]config.UpstreamConfig{})
	router := NewRouter(&config.Config{}, database, r)

	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Upstream
	upBody := `{"key_id":"test-ups","model":"test-model","base_url":"http://test","api_key":"sk-123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/upstreams", bytes.NewBufferString(upBody))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// List Upstreams
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/upstreams", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var upstreams []db.Upstream
	json.Unmarshal(rr.Body.Bytes(), &upstreams)
	assert.Len(t, upstreams, 1)
	assert.Equal(t, "test-ups", upstreams[0].KeyID)

	// Delete Upstream
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/1/upstreams/test-ups", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestNotImplementedAdminHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(&config.Config{}, database, nil)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/tenants/1"},
		{http.MethodGet, "/api/v1/tenants/1/rules"},
		{http.MethodPost, "/api/v1/tenants/1/rules"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotImplemented, rr.Code)
	}
}
