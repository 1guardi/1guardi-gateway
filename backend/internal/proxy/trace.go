package proxy

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const maxAttrLen = 4096

var tracer = otel.Tracer("ai-gateway/proxy")

// agentTraceMiddleware creates agent → thread child spans so every handler
// runs inside: otelhttp("proxy") → agent → thread → operation spans.
// Must be registered after Authenticate (needs populated TenantContext).
func agentTraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := TenantCtx(r.Context())

		ctx, agentSpan := tracer.Start(r.Context(), "agent",
			trace.WithAttributes(
				attribute.String("agent.id", tc.AgentID),
				attribute.String("tenant.id", tc.TenantID),
			),
		)
		defer agentSpan.End()

		ctx, threadSpan := tracer.Start(ctx, "thread",
			trace.WithAttributes(
				attribute.String("thread.id", tc.ThreadID),
				attribute.String("session.id", tc.SessionID),
			),
		)
		defer threadSpan.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// truncate caps string span attributes at maxAttrLen to stay within OTel limits.
func truncate(s string) string {
	if len(s) <= maxAttrLen {
		return s
	}
	return s[:maxAttrLen]
}
