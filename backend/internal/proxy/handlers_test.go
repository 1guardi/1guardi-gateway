package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

// testServer returns a Server with no router configured.
// Validation logic runs before the router is touched, so nil is safe for these tests.
func testServer() *Server {
	return &Server{router: nil, httpClient: &http.Client{}}
}

// serverWithUpstream builds a Server whose router points at the given upstream URL.
func serverWithUpstream(t *testing.T, upstreamURL, model string) *Server {
	t.Helper()
	return &Server{
		router: llmrouter.New([]config.UpstreamConfig{{
			KeyID:   "test",
			Model:   model,
			BaseURL: upstreamURL,
			APIKey:  "test-key",
		}}),
		httpClient: &http.Client{},
	}
}

func chatBody(t *testing.T, extra map[string]interface{}) *bytes.Buffer {
	t.Helper()
	body := map[string]interface{}{
		"model":    "gpt-4o",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	}
	for k, v := range extra {
		body[k] = v
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// ---- validation ----

func TestHandleChatCompletions_Validation(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{invalid`))
		rr := httptest.NewRecorder()

		testServer().handleChatCompletions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp errorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request_error", errResp.Error.Type)
	})

	t.Run("missing model", func(t *testing.T) {
		body := map[string]interface{}{
			"messages": []map[string]string{{"role": "user", "content": "hello"}},
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		testServer().handleChatCompletions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "model is required")
	})

	t.Run("missing messages", func(t *testing.T) {
		body := map[string]interface{}{"model": "gpt-4o"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		testServer().handleChatCompletions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "messages is required")
	})
}

// ---- routing failures ----

func TestHandleChatCompletions_NoUpstream(t *testing.T) {
	srv := &Server{
		router:     llmrouter.New(nil), // no endpoints configured
		httpClient: &http.Client{},
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", chatBody(t, nil))
	rr := httptest.NewRecorder()

	srv.handleChatCompletions(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "no available upstream")
}

// ---- upstream proxy ----

func TestHandleChatCompletions_NonStreaming(t *testing.T) {
	const upstreamResp = `{"id":"chatcmpl-1","choices":[{"message":{"role":"assistant","content":"Hi"}}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, upstreamResp)
	}))
	defer upstream.Close()

	srv := serverWithUpstream(t, upstream.URL, "gpt-4o")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", chatBody(t, nil))
	rr := httptest.NewRecorder()

	srv.handleChatCompletions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, upstreamResp, rr.Body.String())
}

func TestHandleChatCompletions_Streaming(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify stream_options was injected.
		var fields map[string]json.RawMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&fields))
		assert.Contains(t, fields, "stream_options")

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" World\"}}],\"usage\":{\"prompt_tokens\":5,\"completion_tokens\":2}}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()

	srv := serverWithUpstream(t, upstream.URL, "gpt-4o")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", chatBody(t, map[string]interface{}{"stream": true}))
	rr := httptest.NewRecorder()

	srv.handleChatCompletions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "data: ")
	assert.Contains(t, body, "[DONE]")
	assert.Contains(t, body, "Hello")
}

func TestHandleChatCompletions_UpstreamNetError(t *testing.T) {
	// Start then immediately close the upstream to force a connection error.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	upstreamURL := upstream.URL
	upstream.Close()

	srv := serverWithUpstream(t, upstreamURL, "gpt-4o")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", chatBody(t, nil))
	rr := httptest.NewRecorder()

	srv.handleChatCompletions(rr, req)

	assert.Equal(t, http.StatusBadGateway, rr.Code)
	assert.Contains(t, rr.Body.String(), "upstream request failed")
}

func TestHandleChatCompletions_Upstream5xx(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":{"message":"Internal server error","type":"server_error"}}`)
	}))
	defer upstream.Close()

	srv := serverWithUpstream(t, upstream.URL, "gpt-4o")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", chatBody(t, nil))
	rr := httptest.NewRecorder()

	srv.handleChatCompletions(rr, req)

	// Gateway forwards the 5xx as-is.
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "server_error")
}

// ---- proxySSE ----

func TestProxySSE_ParsesUsageAndForwardsChunks(t *testing.T) {
	sseBody := strings.NewReader(
		"data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n" +
			"data: {\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\n" +
			"data: [DONE]\n\n",
	)
	rr := httptest.NewRecorder()

	ttft, input, output := proxySSE(rr, sseBody, time.Now())

	assert.GreaterOrEqual(t, ttft, 0.0)
	assert.Equal(t, 3, input)
	assert.Equal(t, 1, output)

	body := rr.Body.String()
	assert.Contains(t, body, "data: ")
	assert.Contains(t, body, "[DONE]")
	assert.Contains(t, body, "Hi")
}

func TestProxySSE_EmptyBody(t *testing.T) {
	rr := httptest.NewRecorder()
	ttft, input, output := proxySSE(rr, strings.NewReader(""), time.Now())

	assert.Equal(t, 0.0, ttft)
	assert.Equal(t, 0, input)
	assert.Equal(t, 0, output)
}

// ---- calcCost ----

func TestCalcCost_KnownModel(t *testing.T) {
	// gpt-4o: $2.50/M input, $10.00/M output
	// 1000 input + 500 output = 0.0025 + 0.005 = 0.0075
	cost := calcCost("gpt-4o", 1000, 500)
	assert.InDelta(t, 0.0075, cost, 0.0001)
}

func TestCalcCost_UnknownModel(t *testing.T) {
	assert.Equal(t, 0.0, calcCost("unknown-model", 1000, 1000))
}

func TestCalcCost_ZeroTokens(t *testing.T) {
	assert.Equal(t, 0.0, calcCost("gpt-4o", 0, 0))
}

// ---- other handlers ----

func TestHandleCompletions_NotImplemented(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/completions", nil)
	rr := httptest.NewRecorder()
	handleCompletions(rr, req)
	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}

func TestHandleEmbeddings_NotImplemented(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", nil)
	rr := httptest.NewRecorder()
	handleEmbeddings(rr, req)
	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}

func TestHandleListModels_NotImplemented(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()
	handleListModels(rr, req)
	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}
