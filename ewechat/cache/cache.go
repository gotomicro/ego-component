package cache

import (
	"context"
	"time"
)

// Cache interface
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Del(ctx context.Context, key string) (int64, error)
}
