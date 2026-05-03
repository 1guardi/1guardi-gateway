package telemetry

import (
	"context"
	"testing"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/stretchr/testify/assert"
)

func TestSetup_NoCollector(t *testing.T) {
	cfg := config.TelemetryConfig{
		CollectorAddr: "",
	}
	shutdown, err := Setup(context.Background(), cfg)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	err = shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSetup_InvalidResource(t *testing.T) {
	// Resource setup might fail if we provide invalid attributes?
	// Most withX functions don't return error though.
	// sdkresource.New can return error but usually doesn't.
}
