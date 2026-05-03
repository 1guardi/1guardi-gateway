package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleListTraces_RequiresAuth verifies unauthenticated requests are rejected.
func TestHandleListTraces_RequiresAuth(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestHandleGetTraceSpans_RequiresAuth verifies unauthenticated requests are rejected.
func TestHandleGetTraceSpans_RequiresAuth(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces/abc123/spans", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestHandleListTraces_NilClickHouse verifies nil ch returns 200 with empty array
// (ClickHouse unavailable must not crash the admin server).
func TestHandleListTraces_NilClickHouse(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var result []traceRowResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(t, result)
}

// TestHandleGetTraceSpans_NilClickHouse verifies nil ch returns 200 with empty array.
func TestHandleGetTraceSpans_NilClickHouse(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces/abc123def456/spans", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var result []traceSpanResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(t, result)
}

// TestHandleListTraces_LimitParam verifies limit query param is accepted (no error).
func TestHandleListTraces_LimitParam(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces?limit=50", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestHandleListTraces_AgentIDParam verifies agent_id filter param is accepted (no error).
func TestHandleListTraces_AgentIDParam(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	url := fmt.Sprintf("/api/v1/tenants/1/traces?agent_id=%d", 42)
	req := authRequest(t, httptest.NewRequest(http.MethodGet, url, nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestHandleListTraces_ResponseShape verifies the JSON shape matches traceRowResponse
// (all required fields present, even when result set is empty).
func TestHandleListTraces_ResponseShape(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Must decode as an array (not object, not null).
	var raw json.RawMessage
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &raw))
	assert.Equal(t, byte('['), raw[0])
}

// TestHandleGetTraceSpans_ResponseShape verifies the JSON shape matches traceSpanResponse.
func TestHandleGetTraceSpans_ResponseShape(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces/trace-abc/spans", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var raw json.RawMessage
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &raw))
	assert.Equal(t, byte('['), raw[0])
}

// TestHandleListTraces_InvalidLimit tests parsing invalid limit parameter
func TestHandleListTraces_InvalidLimit(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces?limit=invalid", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var result []traceRowResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(t, result)
}

// TestHandleGetTraceSpans_EmptyTraceID tests with empty trace ID
func TestHandleGetTraceSpans_EmptyTraceID(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces//spans", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should handle empty trace ID gracefully
	assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound)
}

// TestHandleListTraces_MultipleParams tests with multiple query parameters
func TestHandleListTraces_MultipleParams(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet,
		"/api/v1/tenants/1/traces?limit=100&agent_id=agent1", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var result []traceRowResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(t, result)
}

// TestHandleListTraces_ZeroLimit tests with limit=0
func TestHandleListTraces_ZeroLimit(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces?limit=0", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

// TestHandleGetTraceSpans_ValidTraceID tests with a valid trace ID
func TestHandleGetTraceSpans_ValidTraceID(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := authRequest(t, httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces/valid-trace-id-123/spans", nil))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	var result []traceSpanResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(t, result)
}

// TestHandleListTraces_NoAuth tests unauthenticated access without Bearer token
func TestHandleListTraces_NoAuth(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces?limit=10", nil)
	// Don't add auth header
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestHandleGetTraceSpans_NoAuth tests unauthenticated access
func TestHandleGetTraceSpans_NoAuth(t *testing.T) {
	router := NewRouter(testCfg(), setupTestDB(t), nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/1/traces/trace-id/spans", nil)
	// Don't add auth header
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
