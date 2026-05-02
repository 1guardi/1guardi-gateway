package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	"github.com/chaitanyabankanhal/ai-gateway/internal/providers"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
	"github.com/redis/go-redis/v9"
)

// Server holds dependencies for the admin API.
type Server struct {
	db        *gorm.DB
	cfg       *config.Config
	llmRouter *llmrouter.Router
	redis     *redis.Client
	modelsSvc *providers.ModelProviderService
}

// NewRouter builds the internal admin API handler.
// This port should never be exposed publicly — bind to 127.0.0.1 in production
// or keep it cluster-internal in Kubernetes.
func NewRouter(cfg *config.Config, database *gorm.DB, llmRouter *llmrouter.Router, redis *redis.Client, modelsSvc *providers.ModelProviderService) http.Handler {
	srv := &Server{
		db:        database,
		cfg:       cfg,
		llmRouter: llmRouter,
		redis:     redis,
		modelsSvc: modelsSvc,
	}

	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.Recoverer)

	// Health + readiness probes — public, no auth required
	mux.Get("/health", handleHealth)
	mux.Get("/ready", srv.handleReady)

	mux.Route("/api/v1", func(r chi.Router) {
		// Public: login
		r.Post("/auth/login", srv.handleLogin)

		// Protected: everything else
		r.Group(func(r chi.Router) {
			r.Use(srv.requireAuth)

			r.Get("/router/endpoints", srv.handleListEndpoints)
			r.Get("/providers/{provider}/models", srv.handleListProviderModels)

			r.Get("/tenants", srv.handleListTenants)
			r.Post("/tenants", srv.handleCreateTenant)

			r.Route("/tenants/{tenantID}", func(r chi.Router) {
				r.Get("/", srv.handleGetTenant)
				r.Patch("/", srv.handleUpdateTenant)
				r.Delete("/", srv.handleDeleteTenant)

				r.Get("/rules", srv.handleListRules)
				r.Post("/rules", srv.handleCreateRule)

				r.Get("/agents", srv.handleListAgents)
				r.Post("/agents", srv.handleCreateAgent)

				r.Get("/keys", srv.handleListKeys)
				r.Post("/keys", srv.handleCreateKey)
				r.Delete("/keys/{keyID}", srv.handleRevokeKey)

				r.Get("/upstreams", srv.handleListUpstreams)
				r.Post("/upstreams", srv.handleCreateUpstream)
				r.Put("/upstreams/{keyID}", srv.handleUpdateUpstream)
				r.Delete("/upstreams/{keyID}", srv.handleDeleteUpstream)
			})
		})
	})

	return mux
}

// requireAuth validates the Bearer JWT on protected routes.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		if _, err := auth.ValidateToken(token, s.cfg.Admin.JWTSecret); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleLogin authenticates an admin user and returns a JWT.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var user db.AdminUser
	if err := s.db.Where("username = ? AND is_active = ?", req.Username, true).First(&user).Error; err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ttl := time.Duration(s.cfg.Admin.JWTTTLHours) * time.Hour
	token, err := auth.GenerateToken(user.ID, user.Username, s.cfg.Admin.JWTSecret, ttl)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      token,
		"expires_at": time.Now().Add(ttl).UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleListEndpoints(w http.ResponseWriter, r *http.Request) {
	if s.llmRouter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.llmRouter.List())
}

func (s *Server) handleListProviderModels(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	apiKey := r.URL.Query().Get("apiKey")
	tenantIDStr := r.URL.Query().Get("tenantID")
	upstreamKeyID := r.URL.Query().Get("upstreamKeyID")

	// Fallback to database key if tenant and key ID are provided
	if apiKey == "" && tenantIDStr != "" && upstreamKeyID != "" {
		tid, _ := strconv.Atoi(tenantIDStr)
		var up db.Upstream
		if err := s.db.Where("tenant_id = ? AND key_id = ?", tid, upstreamKeyID).First(&up).Error; err == nil {
			apiKey = up.APIKey
		}
	}

	// Fallback to default API key from config if still not found
	if apiKey == "" {
		for _, u := range s.cfg.Upstreams {
			if u.Provider == provider && u.APIKey != "" {
				apiKey = u.APIKey
				break
			}
		}
	}

	if apiKey == "" {
		http.Error(w, "API key required and no default found", http.StatusBadRequest)
		return
	}

	models, err := s.modelsSvc.GetModels(r.Context(), provider, apiKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch models: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
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
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	key, hash, suffix, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "failed to generate api key", http.StatusInternalServerError)
		return
	}

	tenant := db.Tenant{
		Name:        req.Name,
		Description: req.Description,
		APIKey:      key,
	}
	if err := s.db.Create(&tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiKey := db.APIKey{
		Name:     "Default Key",
		KeyHash:  hash,
		Prefix:   auth.KeyPrefix,
		Suffix:   suffix,
		TenantID: tenant.ID,
		IsActive: true,
	}
	if err := s.db.Create(&apiKey).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, _ := strconv.Atoi(tenantIDStr)

	var tenant db.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	var agentCount, keyCount int64
	s.db.Model(&db.Agent{}).Where("tenant_id = ?", tenantID).Count(&agentCount)
	s.db.Model(&db.APIKey{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&keyCount)

	resp := struct {
		db.Tenant
		AgentCount int64 `json:"agent_count"`
		KeyCount   int64 `json:"key_count"`
	}{
		Tenant:     tenant,
		AgentCount: agentCount,
		KeyCount:   keyCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, _ := strconv.Atoi(tenantIDStr)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var tenant db.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := s.db.Model(&tenant).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, _ := strconv.Atoi(tenantIDStr)

	var tenant db.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	// Cascade: remove upstreams from live router
	if s.llmRouter != nil {
		var upstreams []db.Upstream
		s.db.Where("tenant_id = ?", tenantID).Find(&upstreams)
		for _, u := range upstreams {
			s.llmRouter.Remove(uint(tenantID), u.KeyID)
		}
	}

	// Soft-delete related records then the tenant
	s.db.Where("tenant_id = ?", tenantID).Delete(&db.Upstream{})
	s.db.Where("tenant_id = ?", tenantID).Delete(&db.APIKey{})
	s.db.Where("tenant_id = ?", tenantID).Delete(&db.Agent{})
	if err := s.db.Delete(&tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	w.WriteHeader(http.StatusCreated)
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

	key, hash, suffix, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "failed to generate key", http.StatusInternalServerError)
		return
	}

	apiKey := db.APIKey{
		Name:     req.Name,
		KeyHash:  hash,
		Prefix:   auth.KeyPrefix,
		Suffix:   suffix,
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
	w.WriteHeader(http.StatusCreated)
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

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleListUpstreams(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.Atoi(chi.URLParam(r, "tenantID"))
	var upstreams []db.Upstream
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&upstreams).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upstreams)
}

func (s *Server) handleCreateUpstream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.Atoi(chi.URLParam(r, "tenantID"))
	var req struct {
		KeyID    string   `json:"key_id"`
		Provider string   `json:"provider"`
		Models   []string `json:"models"`
		Model    string   `json:"model"` // backward compat
		BaseURL  string   `json:"base_url"`
		APIKey   string   `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Provider == "" {
		req.Provider = "openai"
	}

	if len(req.Models) == 0 && req.Model != "" {
		req.Models = []string{req.Model}
	}

	if len(req.Models) == 0 {
		http.Error(w, "at least one model is required", http.StatusBadRequest)
		return
	}

	up := db.Upstream{
		KeyID:    req.KeyID,
		Provider: req.Provider,
		Models:   strings.Join(req.Models, ","),
		BaseURL:  req.BaseURL,
		APIKey:   req.APIKey,
		TenantID: uint(tenantID),
	}

	if err := s.db.Create(&up).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.llmRouter != nil {
		for _, model := range req.Models {
			s.llmRouter.Add(config.UpstreamConfig{
				KeyID:    up.KeyID,
				Provider: up.Provider,
				Model:    model,
				BaseURL:  up.BaseURL,
				APIKey:   up.APIKey,
				TenantID: up.TenantID,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(up)
}

func (s *Server) handleUpdateUpstream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.Atoi(chi.URLParam(r, "tenantID"))
	keyID := chi.URLParam(r, "keyID")

	var req struct {
		Provider string   `json:"provider"`
		Models   []string `json:"models"`
		BaseURL  string   `json:"base_url"`
		APIKey   string   `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var up db.Upstream
	if err := s.db.Where("tenant_id = ? AND key_id = ?", tenantID, keyID).First(&up).Error; err != nil {
		http.Error(w, "upstream not found", http.StatusNotFound)
		return
	}

	if req.Provider != "" {
		up.Provider = req.Provider
	}
	if len(req.Models) > 0 {
		up.Models = strings.Join(req.Models, ",")
	}
	if req.BaseURL != "" {
		up.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		up.APIKey = req.APIKey
	}

	if err := s.db.Save(&up).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.llmRouter != nil {
		s.llmRouter.Remove(uint(tenantID), keyID)
		for _, model := range req.Models {
			s.llmRouter.Add(config.UpstreamConfig{
				KeyID:    up.KeyID,
				Provider: up.Provider,
				Model:    model,
				BaseURL:  up.BaseURL,
				APIKey:   up.APIKey,
				TenantID: up.TenantID,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(up)
}

func (s *Server) handleDeleteUpstream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.Atoi(chi.URLParam(r, "tenantID"))
	keyID := chi.URLParam(r, "keyID")

	var up db.Upstream
	if err := s.db.Where("tenant_id = ? AND key_id = ?", tenantID, keyID).First(&up).Error; err != nil {
		http.Error(w, "upstream not found", http.StatusNotFound)
		return
	}

	if err := s.db.Delete(&up).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.llmRouter != nil {
		s.llmRouter.Remove(uint(tenantID), keyID)
	}

	w.WriteHeader(http.StatusNoContent)
}
