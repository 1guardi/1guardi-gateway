package router

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

var ErrNoEndpoint = errors.New("no available endpoint for model")

// EndpointStatus is a snapshot of one upstream endpoint's live metrics.
type EndpointStatus struct {
	ID           string  `json:"id"`
	Label        string  `json:"label"`
	Model        string  `json:"model"`
	TenantID     uint    `json:"tenant_id"`
	TTFTP50Ms    float64 `json:"ttft_p50_ms"`
	TTFTP99Ms    float64 `json:"ttft_p99_ms"`
	AvgTPS       float64 `json:"avg_tps"`
	ErrorRate    float64 `json:"error_rate"`
	QuotaUsed    float64 `json:"quota_used"`
	CircuitState string  `json:"circuit_state"`
	Score        float64 `json:"score"`
}

// Router selects upstream endpoints based on live performance signals.
// Fallback-only: no load balancing. Highest-scoring available endpoint wins.
type Router struct {
	mu        sync.RWMutex
	endpoints []*Endpoint
}

// New constructs a Router from the configured upstreams.
func New(upstreams []config.UpstreamConfig) *Router {
	endpoints := make([]*Endpoint, len(upstreams))
	for i, u := range upstreams {
		endpoints[i] = newEndpoint(u)
	}
	return &Router{endpoints: endpoints}
}

// List returns a live snapshot of all configured endpoints.
func (r *Router) List() []EndpointStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]EndpointStatus, len(r.endpoints))
	for i, e := range r.endpoints {
		out[i] = EndpointStatus{
			ID:           e.KeyID(),
			Label:        e.KeyID(),
			Model:        e.Model(),
			TenantID:     e.TenantID(),
			TTFTP50Ms:    e.TTFTP50(),
			TTFTP99Ms:    e.TTFTP99(),
			AvgTPS:       e.AvgTPS(),
			ErrorRate:    e.ErrorRate(),
			QuotaUsed:    0,
			CircuitState: e.CircuitStateName(),
			Score:        e.Score(),
		}
	}
	return out
}

// Add appends a new endpoint to the router.
func (r *Router) Add(cfg config.UpstreamConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endpoints = append(r.endpoints, newEndpoint(cfg))
}

// Remove removes an endpoint from the router by KeyID and TenantID.
func (r *Router) Remove(tenantID uint, keyID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var next []*Endpoint
	for _, e := range r.endpoints {
		if e.TenantID() == tenantID && e.KeyID() == keyID {
			continue
		}
		next = append(next, e)
	}
	r.endpoints = next
}

// Pick returns the best available endpoint for the given model and tenant.
// Returns ErrNoEndpoint if all matching endpoints have open circuits or none are configured.
func (r *Router) Pick(tenantID uint, model string) (*Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*Endpoint
	for _, e := range r.endpoints {
		if e.TenantID() == tenantID && e.Model() == model && e.Available() {
			candidates = append(candidates, e)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: %s (tenant %d)", ErrNoEndpoint, model, tenantID)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score() > candidates[j].Score()
	})

	return candidates[0], nil
}
