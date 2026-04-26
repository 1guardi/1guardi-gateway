package proxy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTenantContext(t *testing.T) {
	ctx := context.Background()
	tc := TenantContext{
		TenantID:  "tenant-1",
		AgentID:   "agent-2",
		ThreadID:  "thread-3",
		SpanID:    "span-4",
		SessionID: "session-5",
	}

	newCtx := withTenantContext(ctx, tc)
	
	// Extract the value using the unexported key directly
	val := newCtx.Value(tenantContextKey)
	assert.NotNil(t, val)
	
	extractedTc, ok := val.(TenantContext)
	assert.True(t, ok)
	assert.Equal(t, tc, extractedTc)
}

func TestTenantCtx(t *testing.T) {
	t.Run("present in context", func(t *testing.T) {
		tc := TenantContext{
			TenantID: "tenant-123",
			AgentID:  "agent-456",
		}
		ctx := context.WithValue(context.Background(), tenantContextKey, tc)

		result := TenantCtx(ctx)
		assert.Equal(t, tc, result)
	})

	t.Run("missing from context", func(t *testing.T) {
		ctx := context.Background()

		result := TenantCtx(ctx)
		assert.Equal(t, TenantContext{}, result, "should return zero-value struct when missing")
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), tenantContextKey, "wrong-type")

		result := TenantCtx(ctx)
		assert.Equal(t, TenantContext{}, result, "should return zero-value struct when type asserts fails")
	})
}
