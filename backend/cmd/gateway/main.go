package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/admin"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	"github.com/chaitanyabankanhal/ai-gateway/internal/proxy"
	llmrouter "github.com/chaitanyabankanhal/ai-gateway/internal/router"
	"github.com/chaitanyabankanhal/ai-gateway/internal/telemetry"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Root context — cancelled on SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// OTel — must be set up before any handlers are created
	shutdownTelemetry, err := telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		slog.Error("failed to setup telemetry", "err", err)
		os.Exit(1)
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(flushCtx); err != nil {
			slog.Error("telemetry shutdown error", "err", err)
		}
	}()

	// Database initialization
	database, err := db.Setup(*cfg)
	if err != nil {
		slog.Error("failed to setup database", "err", err)
		os.Exit(1)
	}

	if err := db.SeedDefaultTenant(database, cfg.Upstreams); err != nil {
		slog.Warn("failed to seed default tenant", "err", err)
	}

	// Redis initialization
	redisCache, err := db.RedisSetup(*cfg)
	if err != nil {
		slog.Error("failed to setup redis", "err", err)
		os.Exit(1)
	}

	// Load all upstreams from DB to initialize the router
	var dbUpstreams []db.Upstream
	if err := database.Find(&dbUpstreams).Error; err != nil {
		slog.Error("failed to load upstreams from db", "err", err)
		os.Exit(1)
	}
	upstreamConfigs := make([]config.UpstreamConfig, len(dbUpstreams))
	for i, u := range dbUpstreams {
		upstreamConfigs[i] = config.UpstreamConfig{
			KeyID:    u.KeyID,
			Provider: u.Provider,
			Model:    u.ProviderModel,
			BaseURL:  u.BaseURL,
			APIKey:   u.APIKey,
			TenantID: u.TenantID,
		}
	}

	// Single router instance shared by both servers so admin can observe live metrics.
	router := llmrouter.New(upstreamConfigs)

	// Two HTTP servers: proxy (hot path) and admin (management)
	proxySrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: proxy.NewRouter(cfg, database, redisCache, router),
		// Long write timeout to accommodate streaming LLM responses
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	adminSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AdminPort),
		Handler:      admin.NewRouter(cfg, database, router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("proxy server listening", "addr", proxySrv.Addr)
		if err := proxySrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("proxy server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("admin server listening", "addr", adminSrv.Addr)
		if err := adminSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("admin server: %w", err)
		}
		return nil
	})

	// Shutdown goroutine — waits for signal then drains both servers
	g.Go(func() error {
		<-gCtx.Done()
		slog.Info("shutdown signal received, draining connections")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shut down in parallel — don't wait for one before starting the other
		shutdownG, _ := errgroup.WithContext(shutdownCtx)
		shutdownG.Go(func() error { return proxySrv.Shutdown(shutdownCtx) })
		shutdownG.Go(func() error { return adminSrv.Shutdown(shutdownCtx) })

		if err := shutdownG.Wait(); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		slog.Info("all connections drained, exiting")
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}
