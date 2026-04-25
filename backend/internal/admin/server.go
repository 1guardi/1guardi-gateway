package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
)

// Server holds dependencies for the admin API.
type Server struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewRouter builds the internal admin API handler.
// This port should never be exposed publicly — bind to 127.0.0.1 in production
// or keep it cluster-internal in Kubernetes.
func NewRouter(cfg *config.Config, database *gorm.DB) http.Handler {
	srv := &Server{
		db:  database,
		cfg: cfg,
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	// Health + readiness probes (used by Docker / k8s)
	r.Get("/health", handleHealth)
	r.Get("/ready", srv.handleReady)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/tenants", srv.handleListTenants)
		r.Post("/tenants", srv.handleCreateTenant)

		r.Route("/tenants/{tenantID}", func(r chi.Router) {
			r.Get("/", srv.handleGetTenant)
			r.Get("/rules", srv.handleListRules)
			r.Post("/rules", srv.handleCreateRule)

			r.Get("/agents", srv.handleListAgents)
			r.Post("/agents", srv.handleCreateAgent)

			r.Get("/keys", srv.handleListKeys)
			r.Post("/keys", srv.handleCreateKey)
			r.Delete("/keys/{keyID}", srv.handleRevokeKey)
		})
	})

	return r
}

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleReady is the readiness probe — checks that backing services are reachable.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := s.db.DB()
	if err != nil {
		http.Error(w, "failed to get underlying DB", http.StatusInternalServerError)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		http.Error(w, "database unreachable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{
		Status: "ready",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	var tenants []db.Tenant
	if err := s.db.Find(&tenants).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenants)
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var tenant db.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.db.Create(&tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	var agents []db.Agent
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&agents).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, _ := strconv.ParseUint(tenantIDStr, 10, 32)

	var agent db.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	agent.TenantID = uint(tenantID)
	if err := s.db.Create(&agent).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	var keys []db.APIKey
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&keys).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Don't return hashes in list
	for i := range keys {
		keys[i].KeyHash = ""
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (s *Server) handleCreateKey(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, _ := strconv.ParseUint(tenantIDStr, 10, 32)

	var req struct {
		Name    string `json:"name"`
		AgentID *uint  `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	key, hash, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "failed to generate key", http.StatusInternalServerError)
		return
	}

	apiKey := db.APIKey{
		Name:     req.Name,
		KeyHash:  hash,
		Prefix:   auth.KeyPrefix,
		TenantID: uint(tenantID),
		AgentID:  req.AgentID,
		IsActive: true,
	}

	if err := s.db.Create(&apiKey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the plaintext key ONLY ONCE
	resp := struct {
		db.APIKey
		Key string `json:"key"`
	}{
		APIKey: apiKey,
		Key:    key,
	}
	resp.KeyHash = "" // Hide hash

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRevokeKey(w http.ResponseWriter, r *http.Request) {
	keyID := chi.URLParam(r, "keyID")
	if err := s.db.Model(&db.APIKey{}).Where("id = ?", keyID).Update("is_active", false).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
