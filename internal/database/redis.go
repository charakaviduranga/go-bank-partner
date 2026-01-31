package database

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string, db int) *redis.Client {
	// Accept either raw host:port ("localhost:6379") or full redis URL ("redis://...")
	if strings.HasPrefix(addr, "redis://") {
		opts, err := redis.ParseURL(addr)
		if err == nil {
			opts.DialTimeout = 5 * time.Second
			opts.ReadTimeout = 3 * time.Second
			opts.WriteTimeout = 3 * time.Second
			return redis.NewClient(opts)
		}
		// Fallback: strip scheme if parse fails
		addr = strings.TrimPrefix(addr, "redis://")
	}

	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

func PingRedis(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
