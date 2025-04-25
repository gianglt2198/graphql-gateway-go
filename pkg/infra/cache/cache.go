package cache

import (
	"context"
	"time"
)

type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// TTL operations
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// Pattern operations
	Keys(ctx context.Context, pattern string) ([]string, error)
	Scan(ctx context.Context, pattern string) Iterator

	// Lock operations
	Lock(ctx context.Context, key string, ttl time.Duration) (Lock, error)

	// Health
	Ping(ctx context.Context) error
	Close() error
}

type Iterator interface {
	Next(ctx context.Context) bool
	Key() string
	Value() []byte
	Error() error
}

type Lock interface {
	Unlock(ctx context.Context) error
	Refresh(ctx context.Context, ttl time.Duration) error
}
