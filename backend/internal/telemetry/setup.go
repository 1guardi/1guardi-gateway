package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/chaitanyabankanhal/ai-gateway/config"
)

// Setup initialises the global TracerProvider, LoggerProvider, and TextMapPropagator.
// Call the returned shutdown function (deferred in main) for a clean flush on exit.
// If CollectorAddr is empty, telemetry is a no-op and Setup returns immediately.
func Setup(ctx context.Context, cfg config.TelemetryConfig) (func(context.Context) error, error) {
	if cfg.CollectorAddr == "" {
		return func(context.Context) error { return nil }, nil
	}
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
		sdkresource.WithProcess(),
		sdkresource.WithOS(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTel resource: %w", err)
	}

	// Trace provider
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.CollectorAddr),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		// TODO: switch to ParentBased(TraceIDRatioBased(0.1)) in production
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Log provider — feeds guardrail audit events (and future agent logs) into ClickHouse via otelcol
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.CollectorAddr),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  0, // retry indefinitely
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP log exporter: %w", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), lp.Shutdown(ctx))
	}, nil
}
