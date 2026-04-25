package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
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

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
