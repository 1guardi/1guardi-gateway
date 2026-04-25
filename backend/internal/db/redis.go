package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/redis/go-redis/v9"
)

// RedisSetup initializes the Redis client.
func RedisSetup(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("connected to redis", "addr", cfg.Redis.Addr)
	return rdb, nil
}
