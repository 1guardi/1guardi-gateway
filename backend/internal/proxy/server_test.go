package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTenantContext(t *testing.T) {
	middleware := extractTenantContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantCtx(r.Context())
		assert.Equal(t, "tenant-123", tc.TenantID)
		assert.Equal(t, "agent-456", tc.AgentID)
		assert.Equal(t, "thread-789", tc.ThreadID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("X-Tenant-Id", "tenant-123")
	req.Header.Set("X-Agent-Id", "agent-456")
	req.Header.Set("X-Thread-Id", "thread-789")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
