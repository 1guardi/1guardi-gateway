package proxy

import "context"

type contextKey string

const tenantContextKey contextKey = "tenantContext"

// TenantContext holds the per-request gateway headers extracted from the inbound call.
// It is attached to the request context by the extractTenantContext middleware and
// consumed by handlers, guardrails, telemetry, and the PII pipeline.
type TenantContext struct {
	TenantID  string
	AgentID   string
	ThreadID  string
	SpanID    string
	SessionID string
}

func withTenantContext(ctx context.Context, tc TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextKey, tc)
}

// TenantCtx retrieves the TenantContext from the request context.
// Returns a zero-value struct if not present (e.g. in tests).
func TenantCtx(ctx context.Context) TenantContext {
	if tc, ok := ctx.Value(tenantContextKey).(TenantContext); ok {
		return tc
	}
	return TenantContext{}
}
