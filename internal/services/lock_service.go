package services

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("could not acquire lock")

type RedisLocker struct {
	client *redis.Client
	prefix string
}

func NewRedisLocker(client *redis.Client, prefix string) *RedisLocker {
	return &RedisLocker{
		client: client,
		prefix: prefix,
	}
}

func (l *RedisLocker) Lock(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	ok, err := l.client.SetNX(
		ctx,
		l.prefix+key,
		"1",
		ttl,
	).Result()

	if err != nil {
		return err
	}

	if !ok {
		return ErrLockNotAcquired
	}

	return nil
}

func (l *RedisLocker) Unlock(ctx context.Context, key string) error {
	return l.client.Del(ctx, l.prefix+key).Err()
}
