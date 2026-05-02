package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

const testJWTSecret = "test-secret-for-unit-tests"

func testCfg() *config.Config {
	return &config.Config{
		Admin: config.AdminConfig{
			JWTSecret:   testJWTSecret,
			JWTTTLHours: 1,
		},
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(database))
	return database
}

// authRequest adds a valid JWT Bearer header to req.
func authRequest(t *testing.T, req *http.Request) *http.Request {
	token, err := auth.GenerateToken(1, "Test User", "test-admin@example.com", true, testJWTSecret, time.Hour)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestHandleLogin(t *testing.T) {
	database := setupTestDB(t)
	cfg := testCfg()
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "secret"
	router := NewRouter(cfg, database, nil, nil, nil)

	require.NoError(t, db.SeedSuperAdmin(database, "admin@example.com", "secret"))

	body := `{"email":"admin@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["token"])
	assert.NotEmpty(t, resp["expires_at"])
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	database := setupTestDB(t)
	require.NoError(t, db.SeedSuperAdmin(database, "admin@example.com", "secret"))
	router := NewRouter(testCfg(), database, nil, nil, nil)

	body := `{"email":"admin@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_Missing(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestTenantHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	// Create tenant
	body := `{"name":"Test Tenant"}`
	req := authRequest(t, httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var created db.Tenant
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))
	assert.Equal(t, "Test Tenant", created.Name)
	assert.NotZero(t, created.ID)

	// List tenants
	req = authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tenants []db.Tenant
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tenants))
	assert.Len(t, tenants, 1)
	assert.Equal(t, created.Name, tenants[0].Name)
}

func TestGetTenantHandler(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	// Create via DB directly
	tenant := db.Tenant{Name: "direct", APIKey: "k"}
	database.Create(&tenant)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "direct", resp["Name"])
}

func TestDeleteTenantHandler(t *testing.T) {
	database := setupTestDB(t)
	r := llmrouter.New([]config.UpstreamConfig{})
	router := NewRouter(testCfg(), database, r, nil, nil)

	tenant := db.Tenant{Name: "to-delete", APIKey: "k"}
	database.Create(&tenant)

	req := authRequest(t, httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/1/", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	var count int64
	database.Unscoped().Model(&db.Tenant{}).Where("id = ?", 1).Count(&count)
	assert.Equal(t, int64(1), count) // soft-deleted, still in DB
}

func TestHandleReady(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp healthResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ready", resp.Status)
}

func TestHandleListEndpoints_NilRouter(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/router/endpoints", nil))
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
	handler := NewRouter(testCfg(), database, r, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/router/endpoints", nil))
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
	router := NewRouter(testCfg(), database, nil, nil, nil)

	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Agent
	req := authRequest(t, httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/agents", bytes.NewBufferString(`{"Name":"agent1"}`)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// List Agents
	req = authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/agents", nil))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var agents []db.Agent
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &agents))
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent1", agents[0].Name)
}

func TestKeyHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Key
	req := authRequest(t, httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/keys", bytes.NewBufferString(`{"name":"test-key"}`)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var created struct {
		ID  uint   `json:"ID"`
		Key string `json:"key"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))
	assert.NotEmpty(t, created.Key)

	// List Keys
	req = authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/keys", nil))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var keys []db.APIKey
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &keys))
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key", keys[0].Name)
	assert.NotEmpty(t, keys[0].Suffix)
	assert.Equal(t, 4, len(keys[0].Suffix))
	assert.Empty(t, keys[0].KeyHash, "KeyHash should be hidden")

	// Revoke Key
	req = authRequest(t, httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/1/keys/1", nil))
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
	router := NewRouter(testCfg(), database, r, nil, nil)

	tenant := db.Tenant{Name: "tenant1"}
	database.Create(&tenant)

	// Create Upstream
	req := authRequest(t, httptest.NewRequest(http.MethodPost, "/api/v1/tenants/1/upstreams",
		bytes.NewBufferString(`{"key_id":"test-ups","model":"test-model","base_url":"http://test","api_key":"sk-123"}`)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// List Upstreams
	req = authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/upstreams", nil))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var upstreams []db.Upstream
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &upstreams))
	assert.Len(t, upstreams, 1)
	assert.Equal(t, "test-ups", upstreams[0].KeyID)

	// Update Upstream
	req = authRequest(t, httptest.NewRequest(http.MethodPut, "/api/v1/tenants/1/upstreams/test-ups",
		bytes.NewBufferString(`{"provider":"openai","models":["gpt-4o","gpt-4o-mini"],"base_url":"https://api.openai.com"}`)))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var updated db.Upstream
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &updated))
	assert.Equal(t, "gpt-4o,gpt-4o-mini", updated.Models)

	// Delete Upstream
	req = authRequest(t, httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/1/upstreams/test-ups", nil))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestNotImplementedAdminHandlers(t *testing.T) {
	database := setupTestDB(t)
	router := NewRouter(testCfg(), database, nil, nil, nil)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/tenants/1/rules"},
		{http.MethodPost, "/api/v1/tenants/1/rules"},
	}

	for _, ep := range endpoints {
		req := authRequest(t, httptest.NewRequest(ep.method, ep.path, nil))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotImplemented, rr.Code, "path: %s", ep.path)
	}
}
