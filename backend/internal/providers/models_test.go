package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelProviderService_GetModels_OpenAI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "/v1/models", r.URL.Path)

		resp := struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}{
			Data: []struct {
				ID string `json:"id"`
			}{
				{ID: "gpt-4"},
				{ID: "gpt-3.5-turbo"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	models, err := svc.GetModels(context.Background(), "openai", "test-key", ts.URL)

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"gpt-4", "gpt-3.5-turbo"}, models)
}

func TestModelProviderService_GetModels_Anthropic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-anthropic", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		assert.Equal(t, "/v1/models", r.URL.Path)

		resp := struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}{
			Data: []struct {
				ID string `json:"id"`
			}{
				{ID: "claude-3-opus"},
				{ID: "claude-3-sonnet"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	svc.anthropicBaseURL = ts.URL
	models, err := svc.GetModels(context.Background(), "anthropic", "test-key-anthropic", "")

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"claude-3-opus", "claude-3-sonnet"}, models)
}

func TestModelProviderService_GetModels_Gemini(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-gemini", r.URL.Query().Get("key"))
		assert.Equal(t, "/v1beta/models", r.URL.Path)

		resp := struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}{
			Models: []struct {
				Name string `json:"name"`
			}{
				{Name: "models/gemini-pro"},
				{Name: "models/gemini-ultra"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	svc.geminiBaseURL = ts.URL
	models, err := svc.GetModels(context.Background(), "gemini", "test-key-gemini", "")

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"models/gemini-pro", "models/gemini-ultra"}, models)
}

func TestModelProviderService_UnsupportedProvider(t *testing.T) {
	svc := NewModelProviderService(nil)
	_, err := svc.GetModels(context.Background(), "unsupported", "key", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestModelProviderService_FetchOpenAIModels_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	_, err := svc.GetModels(context.Background(), "openai", "key", ts.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestModelProviderService_FetchOpenAIModels_NonJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	_, err := svc.GetModels(context.Background(), "openai", "key", ts.URL)
	assert.Error(t, err)
}

func TestModelProviderService_FetchGeminiModels_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	svc.geminiBaseURL = ts.URL
	_, err := svc.GetModels(context.Background(), "gemini", "key", "")
	assert.Error(t, err)
}

func TestModelProviderService_FetchAnthropicModels_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	svc := NewModelProviderService(nil)
	svc.anthropicBaseURL = ts.URL
	_, err := svc.GetModels(context.Background(), "anthropic", "key", "")
	assert.Error(t, err)
}
