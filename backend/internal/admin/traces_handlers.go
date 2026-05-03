package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type traceRowResponse struct {
	TraceID      string  `json:"trace_id"`
	Timestamp    string  `json:"ts"`
	Model        string  `json:"model"`
	InputTokens  int32   `json:"input_tokens"`
	OutputTokens int32   `json:"output_tokens"`
	Cost         float64 `json:"cost"`
	TtftMs       float64 `json:"ttft_ms"`
	Tps          float64 `json:"tps"`
	DurationMs   float64 `json:"duration_ms"`
	Status       string  `json:"status"`
	AgentID      string  `json:"agent_id"`
	ThreadID     string  `json:"thread_id"`
}

type traceSpanResponse struct {
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id"`
	SpanName     string            `json:"span_name"`
	DurationMs   float64           `json:"duration_ms"`
	StartTimeMs  float64           `json:"start_time_ms"`
	StatusCode   string            `json:"status_code"`
	Attributes   map[string]string `json:"attributes"`
}

func (s *Server) handleListTraces(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	agentID := r.URL.Query().Get("agent_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if s.ch == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]")) //nolint:errcheck
		return
	}

	rows, err := s.ch.ListTraces(r.Context(), tenantID, agentID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]traceRowResponse, len(rows))
	for i, row := range rows {
		out[i] = traceRowResponse{
			TraceID:      row.TraceID,
			Timestamp:    row.Timestamp,
			Model:        row.Model,
			InputTokens:  row.InputTokens,
			OutputTokens: row.OutputTokens,
			Cost:         row.Cost,
			TtftMs:       row.TtftMs,
			Tps:          row.Tps,
			DurationMs:   row.DurationMs,
			Status:       row.Status,
			AgentID:      row.AgentID,
			ThreadID:     row.ThreadID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out) //nolint:errcheck
}

func (s *Server) handleGetTraceSpans(w http.ResponseWriter, r *http.Request) {
	traceID := chi.URLParam(r, "traceID")

	if s.ch == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]")) //nolint:errcheck
		return
	}

	spans, err := s.ch.GetTraceSpans(r.Context(), traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]traceSpanResponse, len(spans))
	for i, sp := range spans {
		attrs := sp.Attributes
		if attrs == nil {
			attrs = map[string]string{}
		}
		out[i] = traceSpanResponse{
			SpanID:       sp.SpanID,
			ParentSpanID: sp.ParentSpanID,
			SpanName:     sp.SpanName,
			DurationMs:   sp.DurationMs,
			StartTimeMs:  sp.StartTimeMs,
			StatusCode:   sp.StatusCode,
			Attributes:   attrs,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out) //nolint:errcheck
}
