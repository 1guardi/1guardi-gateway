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
	router := NewRouter(cfg, database)

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
	router := NewRouter(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var resp healthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ready", resp.Status)
}
