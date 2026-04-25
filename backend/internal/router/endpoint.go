package router

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

const (
	windowSize    = 50
	failThreshold = 5
	probeInterval = 30 * time.Second

	// Scoring weights per spec: score = w1*(1/TTFT_P99) + w2*avgTPS + w3*(1-errorRate)
	w1 = 0.5
	w2 = 0.3
	w3 = 0.2
)

type circuitState int

const (
	stateClosed   circuitState = iota
	stateOpen
	stateHalfOpen
)

type circuitBreaker struct {
	mu       sync.Mutex
	state    circuitState
	failures int
	openedAt time.Time
}

func (cb *circuitBreaker) available() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Since(cb.openedAt) >= probeInterval {
			cb.state = stateHalfOpen
			return true
		}
		return false
	case stateHalfOpen:
		return true
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == stateHalfOpen || cb.failures >= failThreshold {
		cb.state = stateOpen
		cb.openedAt = time.Now()
		cb.failures = 0
	}
}

// rollingWindow tracks the last N float64 values.
type rollingWindow struct {
	mu     sync.Mutex
	values []float64
	size   int
	pos    int
	count  int
}

func newRollingWindow(size int) *rollingWindow {
	return &rollingWindow{values: make([]float64, size), size: size}
}

func (w *rollingWindow) record(v float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.values[w.pos%w.size] = v
	w.pos++
	if w.count < w.size {
		w.count++
	}
}

func (w *rollingWindow) p99() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count == 0 {
		return 0
	}
	buf := make([]float64, w.count)
	copy(buf, w.values[:w.count])
	sort.Float64s(buf)
	idx := int(math.Ceil(0.99*float64(w.count))) - 1
	if idx < 0 {
		idx = 0
	}
	return buf[idx]
}

func (w *rollingWindow) avg() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < w.count; i++ {
		sum += w.values[i]
	}
	return sum / float64(w.count)
}

type endpointSignals struct {
	ttfts  *rollingWindow // TTFT in ms
	tpss   *rollingWindow // tokens per second
	errors *rollingWindow // 0.0 = success, 1.0 = error
}

func newSignals() *endpointSignals {
	return &endpointSignals{
		ttfts:  newRollingWindow(windowSize),
		tpss:   newRollingWindow(windowSize),
		errors: newRollingWindow(windowSize),
	}
}

// Endpoint wraps an UpstreamConfig with live performance signals and a circuit breaker.
type Endpoint struct {
	cfg     config.UpstreamConfig
	signals *endpointSignals
	cb      *circuitBreaker
}

func newEndpoint(cfg config.UpstreamConfig) *Endpoint {
	return &Endpoint{cfg: cfg, signals: newSignals(), cb: &circuitBreaker{}}
}

func (e *Endpoint) BaseURL() string { return e.cfg.BaseURL }
func (e *Endpoint) APIKey() string  { return e.cfg.APIKey }
func (e *Endpoint) Model() string   { return e.cfg.Model }
func (e *Endpoint) KeyID() string   { return e.cfg.KeyID }

func (e *Endpoint) Available() bool { return e.cb.available() }

func (e *Endpoint) RecordSuccess(ttftMS, tps float64) {
	e.signals.ttfts.record(ttftMS)
	e.signals.tpss.record(tps)
	e.signals.errors.record(0)
	e.cb.recordSuccess()
}

func (e *Endpoint) RecordError() {
	e.signals.errors.record(1)
	e.cb.recordFailure()
}

// Score returns a composite performance score. Higher is better.
// score = w1*(1/TTFT_P99) + w2*avgTPS + w3*(1-errorRate)
// New endpoints with no data score at a neutral 0.5 to attract initial traffic.
func (e *Endpoint) Score() float64 {
	ttftP99 := e.signals.ttfts.p99()
	avgTPS := e.signals.tpss.avg()
	errorRate := e.signals.errors.avg()

	var score float64
	if ttftP99 > 0 {
		score += w1 * (1 / ttftP99)
	} else {
		score += w1 * 0.5
	}
	score += w2 * avgTPS
	score += w3 * (1 - errorRate)
	return score
}
