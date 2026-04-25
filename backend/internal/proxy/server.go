package proxy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
)

// Server holds shared dependencies for the proxy hot path.
type Server struct {
	router     *llmrouter.Router
	httpClient *http.Client
	db         *gorm.DB
}

// NewRouter builds the OpenAI-compatible proxy HTTP handler.
func NewRouter(cfg *config.Config, database *gorm.DB) http.Handler {
	srv := &Server{
		router: llmrouter.New(cfg.Upstreams),
		db:     database,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 20,
			},
			// No hard timeout — streaming responses can be long.
			// Callers control deadline via request context.
		},
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(srv.Authenticate)

	// OpenAI-compatible surface
	r.Post("/v1/chat/completions", srv.handleChatCompletions)
	r.Post("/v1/completions", handleCompletions)
	r.Post("/v1/embeddings", handleEmbeddings)
	r.Get("/v1/models", handleListModels)

	return otelhttp.NewHandler(r, "proxy",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)
}

// Authenticate validates the API key and attaches tenant context.
func (s *Server) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header", "invalid_request_error")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header format", "invalid_request_error")
			return
		}

		key := parts[1]
		hash := auth.HashKey(key)

		var apiKey db.APIKey
		if err := s.db.Where("key_hash = ? AND is_active = ?", hash, true).First(&apiKey).Error; err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or inactive api key", "invalid_request_error")
			return
		}

		// Update last used at
		now := time.Now()
		s.db.Model(&apiKey).Update("last_used_at", &now)

		// Attach tenant and agent info to context.
		// We preserve the ability to pass these in headers for project-level keys.
		tc := TenantContext{
			TenantID:  r.Header.Get("X-Tenant-Id"),
			AgentID:   r.Header.Get("X-Agent-Id"),
			ThreadID:  r.Header.Get("X-Thread-Id"),
			SpanID:    r.Header.Get("X-Span-Id"),
			SessionID: r.Header.Get("X-Session-Id"),
		}

		// Security: TenantID must match the API key's tenant.
		apiKeyTenantID := fmt.Sprintf("%d", apiKey.TenantID)
		if tc.TenantID == "" {
			tc.TenantID = apiKeyTenantID
		} else if tc.TenantID != apiKeyTenantID {
			writeError(w, http.StatusForbidden, "api key is not authorized for this tenant", "invalid_request_error")
			return
		}

		// Security: If key is scoped to an agent, AgentID must match or be empty.
		if apiKey.AgentID != nil {
			scopedAgentID := fmt.Sprintf("%d", *apiKey.AgentID)
			if tc.AgentID != "" && tc.AgentID != scopedAgentID {
				writeError(w, http.StatusForbidden, "api key is not authorized for this agent", "invalid_request_error")
				return
			}
			tc.AgentID = scopedAgentID
		}

		next.ServeHTTP(w, r.WithContext(withTenantContext(r.Context(), tc)))
	})
}
