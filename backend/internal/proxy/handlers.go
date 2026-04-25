package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

const maxBodyBytes = 10 << 20 // 10MB

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

// streamChunk is the minimal shape of an OpenAI SSE chunk used for usage extraction.
type streamChunk struct {
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

// chatResponse is the minimal shape of a non-streaming OpenAI response.
type chatResponse struct {
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// modelPrice holds per-million-token USD pricing.
type modelPrice struct {
	inputPerMTok  float64
	outputPerMTok float64
}

var pricingTable = map[string]modelPrice{
	"gpt-4o":                     {2.50, 10.00},
	"gpt-4o-mini":                {0.15, 0.60},
	"gpt-4-turbo":                {10.00, 30.00},
	"gpt-3.5-turbo":              {0.50, 1.50},
	"claude-opus-4-7":            {15.00, 75.00},
	"claude-sonnet-4-6":          {3.00, 15.00},
	"claude-haiku-4-5-20251001":  {0.80, 4.00},
}

func calcCost(model string, inputTokens, outputTokens int) float64 {
	p, ok := pricingTable[model]
	if !ok {
		return 0
	}
	return (float64(inputTokens)*p.inputPerMTok + float64(outputTokens)*p.outputPerMTok) / 1_000_000
}

// handleChatCompletions handles POST /v1/chat/completions.
// Pipeline:
//  1. Parse + validate request
//  2. Route to upstream LLM via router (with circuit breaker)
//  3. Stream response back, tracking TTFT + TPS
//  4. Emit OTel span with gen_ai.* attributes
//
// Guardrails and PII masking are deferred to subsequent milestones.
func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	// 1. Read and parse raw body so we can both validate and forward.
	rawBody, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body", "invalid_request_error")
		return
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &fields); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", "invalid_request_error")
		return
	}

	// Extract model (required).
	var model string
	if raw, ok := fields["model"]; ok {
		if err := json.Unmarshal(raw, &model); err != nil || model == "" {
			writeError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
			return
		}
	} else {
		writeError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
		return
	}

	// Validate messages present.
	if _, ok := fields["messages"]; !ok {
		writeError(w, http.StatusBadRequest, "messages is required", "invalid_request_error")
		return
	}

	// Extract stream flag.
	var streaming bool
	if raw, ok := fields["stream"]; ok {
		json.Unmarshal(raw, &streaming) //nolint:errcheck — bool unmarshal never errors on valid JSON
	}

	// 2. Pick upstream endpoint.
	endpoint, err := s.router.Pick(model)
	if err != nil {
		if errors.Is(err, llmrouter.ErrNoEndpoint) {
			writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("no available upstream for model %q", model), "api_error")
		} else {
			writeError(w, http.StatusInternalServerError, "router error", "api_error")
		}
		return
	}

	// 3. If streaming, inject stream_options.include_usage so we get token counts
	//    in the final SSE chunk without extra requests.
	forwardBody := rawBody
	if streaming {
		fields["stream_options"] = json.RawMessage(`{"include_usage":true}`)
		if forwardBody, err = json.Marshal(fields); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to marshal request", "api_error")
			return
		}
	}

	// 4. Build and send upstream request.
	upstreamURL := strings.TrimRight(endpoint.BaseURL(), "/") + "/v1/chat/completions"
	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(forwardBody))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build upstream request", "api_error")
		return
	}
	upstreamReq.Header.Set("Authorization", "Bearer "+endpoint.APIKey())
	upstreamReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		endpoint.RecordError()
		writeError(w, http.StatusBadGateway, "upstream request failed", "api_error")
		return
	}
	defer resp.Body.Close()

	// 5. Copy upstream response headers.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 6. Stream or buffer the response, collecting telemetry.
	var (
		ttftMS       float64
		inputTokens  int
		outputTokens int
	)

	if streaming {
		ttftMS, inputTokens, outputTokens = proxySSE(w, resp.Body, start)
	} else {
		body, _ := io.ReadAll(resp.Body)
		ttftMS = float64(time.Since(start).Milliseconds())
		var result chatResponse
		if json.Unmarshal(body, &result) == nil {
			inputTokens = result.Usage.PromptTokens
			outputTokens = result.Usage.CompletionTokens
		}
		w.Write(body) //nolint:errcheck — client disconnect is non-fatal
	}

	// 7. Record signals and emit OTel attributes.
	totalSec := time.Since(start).Seconds()
	var tps float64
	if totalSec > 0 && outputTokens > 0 {
		tps = float64(outputTokens) / totalSec
	}

	if resp.StatusCode >= 500 {
		endpoint.RecordError()
	} else {
		endpoint.RecordSuccess(ttftMS, tps)
	}

	tc := TenantCtx(r.Context())
	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(
		attribute.String("gen_ai.model", model),
		attribute.String("gen_ai.thread.id", tc.ThreadID),
		attribute.String("gen_ai.agent.id", tc.AgentID),
		attribute.Int("gen_ai.input.tokens", inputTokens),
		attribute.Int("gen_ai.output.tokens", outputTokens),
		attribute.Float64("gen_ai.ttft_ms", ttftMS),
		attribute.Float64("gen_ai.tps", tps),
		attribute.Float64("gen_ai.cost.usd", calcCost(model, inputTokens, outputTokens)),
	)
}

// proxySSE streams an SSE response to w, returning TTFT ms and token counts from
// the final usage chunk.
func proxySSE(w http.ResponseWriter, body io.Reader, start time.Time) (ttftMS float64, inputTokens, outputTokens int) {
	const maxSSELine = 1 << 20 // 1MB per line max
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, maxSSELine), maxSSELine)

	flusher, canFlush := w.(http.Flusher)
	var firstToken bool

	for scanner.Scan() {
		line := scanner.Text()

		// Record TTFT on the first data line.
		if !firstToken && strings.HasPrefix(line, "data: ") {
			ttftMS = float64(time.Since(start).Milliseconds())
			firstToken = true
		}

		fmt.Fprintf(w, "%s\n", line)

		// Flush on event boundary (empty line separates SSE events).
		if line == "" && canFlush {
			flusher.Flush()
		}

		// Parse usage from every chunk — OpenAI populates it only on the final one.
		if strings.HasPrefix(line, "data: ") && line != "data: [DONE]" {
			data := line[len("data: "):]
			var chunk streamChunk
			if json.Unmarshal([]byte(data), &chunk) == nil && chunk.Usage != nil {
				inputTokens = chunk.Usage.PromptTokens
				outputTokens = chunk.Usage.CompletionTokens
			}
		}
	}

	return ttftMS, inputTokens, outputTokens
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
