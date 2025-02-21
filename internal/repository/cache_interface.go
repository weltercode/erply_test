package cache

import (
	"context"
	"time"
)

type CacheInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Close() error
}
