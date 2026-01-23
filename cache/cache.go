package cache

import (
	"context"
	"time"
)

// Cache is the interface for response caching
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, bool)
	
	// Set stores a value in the cache with a TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	
	// Clear removes all values from the cache
	Clear(ctx context.Context) error
	
	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits   uint64
	Misses uint64
	Size   uint64
	Items  uint64
}

// Config holds cache configuration
type Config struct {
	// MaxSize is the maximum cache size in bytes (default: 100MB)
	MaxSize int64
	
	// MaxItems is the maximum number of items (default: 10000)
	MaxItems int
	
	// DefaultTTL is the default TTL for cached items (default: 5 minutes)
	DefaultTTL time.Duration
	
	// Enabled indicates whether caching is enabled
	Enabled bool
}

// DefaultConfig returns a default cache configuration
func DefaultConfig() *Config {
	return &Config{
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxItems:   10000,
		DefaultTTL: 5 * time.Minute,
		Enabled:    true,
	}
}
