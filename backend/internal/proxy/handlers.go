package proxy

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the OpenAI-compatible error envelope.
type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := errorResponse{}
	resp.Error.Message = msg
	resp.Error.Type = errType
	json.NewEncoder(w).Encode(resp)
}

// handleChatCompletions handles POST /v1/chat/completions.
// Pipeline (to be implemented incrementally):
//  1. Parse + validate request
//  2. Run inbound guardrails (parallel)
//  3. PII masking on input
//  4. Route to upstream LLM via router (with circuit breaker)
//  5. Stream response back, tracking TTFT + TPS
//  6. Run outbound guardrails on response
//  7. PII unmasking on output
//  8. Emit OTel span with gen_ai.* attributes
func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented", "api_error")
}

// handleCompletions handles POST /v1/completions (legacy).
func handleCompletions(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented", "api_error")
}

// handleEmbeddings handles POST /v1/embeddings.
func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented", "api_error")
}

// handleListModels handles GET /v1/models.
func handleListModels(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented", "api_error")
}
