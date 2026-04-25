package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleChatCompletions_Validation(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{invalid`))
		rr := httptest.NewRecorder()

		handleChatCompletions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		
		var errResp errorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request_error", errResp.Error.Type)
	})

	t.Run("missing model", func(t *testing.T) {
		body := map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": "hello"},
			},
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		handleChatCompletions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "model is required")
	})
}
