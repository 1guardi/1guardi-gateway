package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

func TestHandleListModels(t *testing.T) {
	// Setup router with some endpoints for tenant 1 and 2
	router := llmrouter.New([]config.UpstreamConfig{
		{KeyID: "k1", TenantID: 1, Model: "gpt-4o", Provider: "openai", BaseURL: "http://localhost", APIKey: "sk-1"},
		{KeyID: "k2", TenantID: 1, Model: "claude-3-opus", Provider: "anthropic", BaseURL: "http://localhost", APIKey: "sk-2"},
		{KeyID: "k3", TenantID: 2, Model: "gpt-3.5-turbo", Provider: "openai", BaseURL: "http://localhost", APIKey: "sk-3"},
	})

	srv := &Server{router: router}

	t.Run("list models for tenant 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		tc := TenantContext{TenantID: "1"}
		ctx := context.WithValue(req.Context(), tenantContextKey, tc)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		srv.handleListModels(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp listModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "list", resp.Object)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "claude-3-opus", resp.Data[0].ID)
		assert.Equal(t, "gpt-4o", resp.Data[1].ID)
	})

	t.Run("list models for tenant 2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		tc := TenantContext{TenantID: "2"}
		ctx := context.WithValue(req.Context(), tenantContextKey, tc)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		srv.handleListModels(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp listModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "gpt-3.5-turbo", resp.Data[0].ID)
	})

	t.Run("list models for tenant with no endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		tc := TenantContext{TenantID: "3"}
		ctx := context.WithValue(req.Context(), tenantContextKey, tc)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		srv.handleListModels(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp listModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Empty(t, resp.Data)
	})
}
