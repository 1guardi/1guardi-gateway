package router

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

// ---- circuit breaker ----

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := &circuitBreaker{}
	assert.True(t, cb.available())
}

func TestCircuitBreaker_OpenAfterThresholdFailures(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold; i++ {
		cb.recordFailure()
	}
	assert.False(t, cb.available())
}

func TestCircuitBreaker_PartialFailuresStayClosed(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold-1; i++ {
		cb.recordFailure()
	}
	assert.True(t, cb.available())
}

func TestCircuitBreaker_OpenToHalfOpenAfterProbeInterval(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold; i++ {
		cb.recordFailure()
	}
	// Simulate probe interval elapsed by backdating openedAt.
	cb.mu.Lock()
	cb.openedAt = time.Now().Add(-(probeInterval + time.Second))
	cb.mu.Unlock()

	assert.True(t, cb.available()) // transitions to HalfOpen and returns true
}

func TestCircuitBreaker_HalfOpenClosedOnSuccess(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold; i++ {
		cb.recordFailure()
	}
	cb.mu.Lock()
	cb.openedAt = time.Now().Add(-(probeInterval + time.Second))
	cb.mu.Unlock()
	cb.available() // trigger HalfOpen transition

	cb.recordSuccess()

	assert.True(t, cb.available())
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()
	assert.Equal(t, stateClosed, state)
}

func TestCircuitBreaker_HalfOpenReopensOnFailure(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold; i++ {
		cb.recordFailure()
	}
	cb.mu.Lock()
	cb.openedAt = time.Now().Add(-(probeInterval + time.Second))
	cb.mu.Unlock()
	cb.available() // trigger HalfOpen transition

	cb.recordFailure()

	assert.False(t, cb.available())
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()
	assert.Equal(t, stateOpen, state)
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := &circuitBreaker{}
	for i := 0; i < failThreshold-1; i++ {
		cb.recordFailure()
	}
	cb.recordSuccess()

	// Need a full failThreshold failures again to open.
	for i := 0; i < failThreshold-1; i++ {
		cb.recordFailure()
	}
	assert.True(t, cb.available())
}

func TestCircuitBreaker_StateNames(t *testing.T) {
	cb := &circuitBreaker{}
	assert.Equal(t, "CLOSED", cb.stateName())

	cb.mu.Lock()
	cb.state = stateOpen
	cb.mu.Unlock()
	assert.Equal(t, "OPEN", cb.stateName())

	cb.mu.Lock()
	cb.state = stateHalfOpen
	cb.mu.Unlock()
	assert.Equal(t, "HALF-OPEN", cb.stateName())

	cb.mu.Lock()
	cb.state = 99 // unknown
	cb.mu.Unlock()
	assert.Equal(t, "UNKNOWN", cb.stateName())
}

// ---- rolling window ----

func TestRollingWindow_EmptyP99(t *testing.T) {
	w := newRollingWindow(10)
	assert.Equal(t, 0.0, w.p99())
}

func TestRollingWindow_EmptyAvg(t *testing.T) {
	w := newRollingWindow(10)
	assert.Equal(t, 0.0, w.avg())
}

func TestRollingWindow_SingleValue(t *testing.T) {
	w := newRollingWindow(10)
	w.record(42.0)
	assert.Equal(t, 42.0, w.p99())
	assert.Equal(t, 42.0, w.avg())
}

func TestRollingWindow_P99(t *testing.T) {
	w := newRollingWindow(100)
	for i := 1; i <= 100; i++ {
		w.record(float64(i))
	}
	// p99 of 1..100: idx = ceil(0.99*100)-1 = 98 → sorted[98] = 99
	assert.Equal(t, 99.0, w.p99())
}

func TestRollingWindow_AvgTwoValues(t *testing.T) {
	w := newRollingWindow(4)
	w.record(1)
	w.record(3)
	assert.Equal(t, 2.0, w.avg())
}

func TestRollingWindow_WrapAround(t *testing.T) {
	// size=3; after 4 records the oldest entry is evicted.
	w := newRollingWindow(3)
	w.record(1)
	w.record(2)
	w.record(3)
	w.record(4) // evicts 1; window = {4, 2, 3}

	assert.Equal(t, 3.0, w.avg()) // (4+2+3)/3
}

// ---- endpoint ----

func TestEndpoint_AccessorsReturnConfigValues(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{
		KeyID:   "k1",
		Model:   "gpt-4o",
		BaseURL: "https://api.openai.com",
		APIKey:  "sk-abc",
	})
	assert.Equal(t, "k1", e.KeyID())
	assert.Equal(t, "gpt-4o", e.Model())
	assert.Equal(t, "https://api.openai.com", e.BaseURL())
	assert.Equal(t, "sk-abc", e.APIKey())
}

func TestEndpoint_AvailableInitially(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{Model: "gpt-4o"})
	assert.True(t, e.Available())
}

func TestEndpoint_NotAvailableAfterErrors(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{Model: "gpt-4o"})
	for i := 0; i < failThreshold; i++ {
		e.RecordError()
	}
	assert.False(t, e.Available())
}

func TestEndpoint_ScoreNoData(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{Model: "gpt-4o"})
	// w1*0.5 + w2*0 + w3*1.0 = 0.25 + 0 + 0.2 = 0.45
	assert.InDelta(t, 0.45, e.Score(), 0.001)
}

func TestEndpoint_ScoreWithData(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{Model: "gpt-4o"})
	e.RecordSuccess(100, 50) // 100ms TTFT, 50 TPS
	// 0.5*(1/100) + 0.3*50 + 0.2*1 = 0.005 + 15 + 0.2 = 15.205
	assert.InDelta(t, 15.205, e.Score(), 0.01)
}

func TestEndpoint_RecordErrorUpdatesSignals(t *testing.T) {
	e := newEndpoint(config.UpstreamConfig{Model: "gpt-4o"})
	e.RecordError()
	// error rate = 1.0 → w3*(1-1) = 0
	score := e.Score()
	// 0.5*0.5 + 0.3*0 + 0.2*0 = 0.25
	assert.InDelta(t, 0.25, score, 0.001)
}

// ---- router ----

func TestRouter_Pick_NoEndpoints(t *testing.T) {
	r := New(nil)
	_, err := r.Pick(0, "gpt-4o")
	assert.ErrorIs(t, err, ErrNoEndpoint)
}

func TestRouter_Pick_NoMatchingModel(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "k1", Model: "gpt-4o", BaseURL: "http://x", APIKey: "k"},
	})
	_, err := r.Pick(0, "gpt-4o-mini")
	assert.ErrorIs(t, err, ErrNoEndpoint)
}

func TestRouter_Pick_ReturnsEndpoint(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "k1", Model: "gpt-4o", BaseURL: "http://x", APIKey: "k"},
	})
	ep, err := r.Pick(0, "gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", ep.Model())
	assert.Equal(t, "k1", ep.KeyID())
}

func TestRouter_Pick_PrefersBetterScore(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "slow", Model: "gpt-4o", BaseURL: "http://slow", APIKey: "k"},
		{KeyID: "fast", Model: "gpt-4o", BaseURL: "http://fast", APIKey: "k"},
	})
	r.endpoints[0].RecordSuccess(500, 20) // slow: high TTFT, low TPS
	r.endpoints[1].RecordSuccess(50, 80)  // fast: low TTFT, high TPS

	ep, err := r.Pick(0, "gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "fast", ep.KeyID())
}

func TestRouter_Pick_SkipsOpenCircuit(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "broken", Model: "gpt-4o", BaseURL: "http://broken", APIKey: "k"},
		{KeyID: "healthy", Model: "gpt-4o", BaseURL: "http://healthy", APIKey: "k"},
	})
	for i := 0; i < failThreshold; i++ {
		r.endpoints[0].RecordError()
	}

	ep, err := r.Pick(0, "gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "healthy", ep.KeyID())
}

func TestRouter_Pick_AllCircuitsOpen(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "k1", Model: "gpt-4o", BaseURL: "http://x", APIKey: "k"},
	})
	for i := 0; i < failThreshold; i++ {
		r.endpoints[0].RecordError()
	}
	_, err := r.Pick(0, "gpt-4o")
	assert.ErrorIs(t, err, ErrNoEndpoint)
}

func TestRouter_List_Empty(t *testing.T) {
	r := New(nil)
	assert.Empty(t, r.List())
}

func TestRouter_List_ReturnsAllEndpoints(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "ep-a", Provider: "openai", Model: "gpt-4o", BaseURL: "http://a", APIKey: "k1"},
		{KeyID: "ep-b", Provider: "anthropic", Model: "claude-3", BaseURL: "http://b", APIKey: "k2"},
	})
	r.endpoints[0].RecordSuccess(80, 50)
	r.endpoints[1].RecordSuccess(200, 30)

	list := r.List()
	require.Len(t, list, 2)

	a := list[0]
	assert.Equal(t, "ep-a", a.ID)
	assert.Equal(t, "openai", a.Provider)
	assert.Equal(t, "gpt-4o", a.Model)
	assert.Equal(t, "CLOSED", a.CircuitState)
	assert.InDelta(t, 80.0, a.TTFTP50Ms, 0.01)
	assert.InDelta(t, 80.0, a.TTFTP99Ms, 0.01)
	assert.InDelta(t, 50.0, a.AvgTPS, 0.01)
	assert.InDelta(t, 0.0, a.ErrorRate, 0.01)
	assert.Positive(t, a.Score)
}

func TestRouter_List_OpenCircuit(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "broken", Model: "gpt-4o", BaseURL: "http://x", APIKey: "k"},
	})
	for i := 0; i < failThreshold; i++ {
		r.endpoints[0].RecordError()
	}

	list := r.List()
	require.Len(t, list, 1)
	assert.Equal(t, "OPEN", list[0].CircuitState)
}

func TestRouter_Add(t *testing.T) {
	r := New([]config.UpstreamConfig{})
	assert.Len(t, r.List(), 0)

	r.Add(config.UpstreamConfig{KeyID: "new-ep", Model: "gpt-4o", BaseURL: "http://test", APIKey: "k"})

	list := r.List()
	require.Len(t, list, 1)
	assert.Equal(t, "new-ep", list[0].ID)
}

func TestRouter_Remove(t *testing.T) {
	r := New([]config.UpstreamConfig{
		{KeyID: "ep1", TenantID: 1, Model: "gpt-4o", BaseURL: "http://test", APIKey: "k"},
		{KeyID: "ep2", TenantID: 1, Model: "gpt-4o", BaseURL: "http://test", APIKey: "k"},
	})

	assert.Len(t, r.List(), 2)

	// Remove matching
	r.Remove(1, "ep1")
	list := r.List()
	require.Len(t, list, 1)
	assert.Equal(t, "ep2", list[0].ID)

	// Remove non-matching (wrong tenant)
	r.Remove(2, "ep2")
	require.Len(t, r.List(), 1)

	// Remove matching
	r.Remove(1, "ep2")
	require.Len(t, r.List(), 0)
}
