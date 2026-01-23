package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter is the interface for rate limiting
type Limiter interface {
	// Allow checks if a request is allowed for the given key
	// Returns true if allowed, false if rate limit exceeded
	Allow(ctx context.Context, key string) bool
	
	// AllowN checks if N requests are allowed for the given key
	AllowN(ctx context.Context, key string, n int) bool
	
	// Reset resets the rate limiter for the given key
	Reset(ctx context.Context, key string)
}

// Config holds rate limiter configuration
type Config struct {
	// RequestsPerSecond is the number of requests allowed per second
	RequestsPerSecond float64
	
	// Burst is the maximum burst size
	Burst int
	
	// Enabled indicates whether rate limiting is enabled
	Enabled bool
}

// DefaultConfig returns a default rate limiter configuration
func DefaultConfig() *Config {
	return &Config{
		RequestsPerSecond: 100, // 100 requests per second
		Burst:             200, // Allow burst of 200
		Enabled:           true,
	}
}

// tokenBucket implements the token bucket rate limiting algorithm
type tokenBucket struct {
	mu     sync.RWMutex
	config *Config
	
	// buckets maps keys to their token buckets
	buckets map[string]*bucket
	
	// cleanupInterval is how often to clean up expired buckets
	cleanupInterval time.Duration
	
	// lastCleanup is the last time cleanup was performed
	lastCleanup time.Time
}

// bucket represents a single token bucket
type bucket struct {
	tokens         float64
	lastRefillTime time.Time
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(config *Config) Limiter {
	if config == nil {
		config = DefaultConfig()
	}
	
	tb := &tokenBucket{
		config:          config,
		buckets:         make(map[string]*bucket),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
	
	return tb
}

// Allow checks if a request is allowed
func (tb *tokenBucket) Allow(ctx context.Context, key string) bool {
	return tb.AllowN(ctx, key, 1)
}

// AllowN checks if N requests are allowed
func (tb *tokenBucket) AllowN(ctx context.Context, key string, n int) bool {
	if !tb.config.Enabled {
		return true
	}
	
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	// Periodic cleanup of old buckets
	if time.Since(tb.lastCleanup) > tb.cleanupInterval {
		tb.cleanup()
	}
	
	// Get or create bucket
	b, exists := tb.buckets[key]
	if !exists {
		b = &bucket{
			tokens:         float64(tb.config.Burst),
			lastRefillTime: time.Now(),
		}
		tb.buckets[key] = b
	}
	
	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastRefillTime).Seconds()
	tokensToAdd := elapsed * tb.config.RequestsPerSecond
	b.tokens = min(b.tokens+tokensToAdd, float64(tb.config.Burst))
	b.lastRefillTime = now
	
	// Check if enough tokens available
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}
	
	return false
}

// Reset resets the rate limiter for a key
func (tb *tokenBucket) Reset(ctx context.Context, key string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	delete(tb.buckets, key)
}

// cleanup removes buckets that haven't been used recently
func (tb *tokenBucket) cleanup() {
	now := time.Now()
	threshold := 10 * time.Minute
	
	for key, b := range tb.buckets {
		if now.Sub(b.lastRefillTime) > threshold {
			delete(tb.buckets, key)
		}
	}
	
	tb.lastCleanup = now
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
