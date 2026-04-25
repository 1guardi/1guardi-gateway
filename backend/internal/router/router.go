package router

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

var ErrNoEndpoint = errors.New("no available endpoint for model")

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

// Pick returns the best available endpoint for the given model.
// Returns ErrNoEndpoint if all matching endpoints have open circuits or none are configured.
func (r *Router) Pick(model string) (*Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*Endpoint
	for _, e := range r.endpoints {
		if e.Model() == model && e.Available() {
			candidates = append(candidates, e)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoEndpoint, model)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score() > candidates[j].Score()
	})

	return candidates[0], nil
}
