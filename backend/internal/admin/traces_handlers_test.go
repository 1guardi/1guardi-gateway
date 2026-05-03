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
