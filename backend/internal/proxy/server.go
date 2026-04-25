package proxy

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

// NewRouter builds the OpenAI-compatible proxy HTTP handler.
// This is the hot path — keep middleware lean.
func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(extractTenantContext)

	// OpenAI-compatible surface
	r.Post("/v1/chat/completions", handleChatCompletions)
	r.Post("/v1/completions", handleCompletions)
	r.Post("/v1/embeddings", handleEmbeddings)
	r.Get("/v1/models", handleListModels)

	// Wrap the whole router so every route gets an OTel span automatically.
	// The span name is set per-route via otelhttp.WithRouteTag in each handler.
	return otelhttp.NewHandler(r, "proxy",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)
}

// extractTenantContext reads the gateway-specific headers and attaches them to
// the request context so downstream middleware and handlers can read them
// without touching raw headers.
func extractTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantContext{
			TenantID:  r.Header.Get("X-Tenant-Id"),
			AgentID:   r.Header.Get("X-Agent-Id"),
			ThreadID:  r.Header.Get("X-Thread-Id"),
			SpanID:    r.Header.Get("X-Span-Id"),
			SessionID: r.Header.Get("X-Session-Id"),
		}
		next.ServeHTTP(w, r.WithContext(withTenantContext(r.Context(), tc)))
	})
}
