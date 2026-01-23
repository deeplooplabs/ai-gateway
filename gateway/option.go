package gateway

import (
	"time"

	"github.com/deeplooplabs/ai-gateway/cache"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/ratelimit"
)

// CORSConfig represents the CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of origins allowed to make requests.
	// Use "*" to allow any origin.
	AllowedOrigins []string
	// AllowedMethods is a list of HTTP methods allowed.
	// If empty, allows all methods.
	AllowedMethods []string
	// AllowedHeaders is a list of headers allowed in requests.
	// If empty, allows all headers.
	AllowedHeaders []string
	// ExposedHeaders is a list of headers exposed to clients.
	ExposedHeaders []string
	// MaxAge is the maximum age to cache preflight responses.
	// If zero, uses default (1 hour).
	MaxAge time.Duration
	// AllowCredentials indicates whether requests can include credentials.
	AllowCredentials bool
}

// DefaultCORSConfig returns a default CORS configuration
// that allows all origins, methods, and headers.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   nil,
		MaxAge:           time.Hour,
		AllowCredentials: false,
	}
}

// Option configures the Gateway
type Option func(*Gateway)

// WithModelRegistry sets the model registry
func WithModelRegistry(registry model.ModelRegistry) Option {
	return func(g *Gateway) {
		g.modelRegistry = registry
	}
}

// WithHooks sets the hook registry
func WithHooks(hooks *hook.Registry) Option {
	return func(g *Gateway) {
		g.hooks = hooks
	}
}

// WithHook registers a single hook
func WithHook(h hook.Hook) Option {
	return func(g *Gateway) {
		g.hooks.Register(h)
	}
}

// WithCORS sets the CORS configuration.
// Pass nil to disable CORS.
func WithCORS(cors *CORSConfig) Option {
	return func(g *Gateway) {
		g.cors = cors
	}
}

// WithMetrics enables Prometheus metrics collection
func WithMetrics(namespace string) Option {
	return func(g *Gateway) {
		g.metrics = NewMetrics(namespace)
	}
}

// WithCache enables response caching
func WithCache(cacheImpl cache.Cache) Option {
	return func(g *Gateway) {
		g.cache = cacheImpl
	}
}

// WithRateLimiter enables rate limiting
func WithRateLimiter(limiter ratelimit.Limiter) Option {
	return func(g *Gateway) {
		g.rateLimiter = limiter
	}
}
