package storage

import "context"

type Set interface {
	Add(ctx context.Context, key string, ttl int64) error
	TryAdd(ctx context.Context, key string, ttl int64) (bool, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, keys []string) (map[string]bool, error)
	// StoreBatch stores a map[string]float64 as a Redis hash with TTL.
	StoreBatch(ctx context.Context, key string, data map[string]float64, ttlSeconds int64) error
	// ReadBatch reads the hash at key. TTL handles expiry.
	ReadBatch(ctx context.Context, key string) (map[string]float64, error)
}
