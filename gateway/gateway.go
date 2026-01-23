package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/deeplooplabs/ai-gateway/cache"
	"github.com/deeplooplabs/ai-gateway/handler"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Gateway is the main HTTP handler
type Gateway struct {
	modelRegistry model.ModelRegistry
	hooks         *hook.Registry
	mux           *http.ServeMux
	cors          *CORSConfig
	metrics       *Metrics
	cache         cache.Cache
	rateLimiter   ratelimit.Limiter
}

// New creates a new gateway with default options
func New(opts ...Option) *Gateway {
	// Default configuration
	g := &Gateway{
		modelRegistry: model.NewMapModelRegistry(),
		hooks:         hook.NewRegistry(),
		mux:           http.NewServeMux(),
	}

	// Apply options
	for _, opt := range opts {
		opt(g)
	}

	// Setup routes
	g.setupRoutes()

	return g
}

func (g *Gateway) setupRoutes() {
	// OpenResponses endpoint
	responsesHandler := handler.NewResponsesHandler(g.modelRegistry, g.hooks)
	g.mux.HandleFunc("/v1/responses", responsesHandler.ServeHTTP)

	// Chat Completions (OpenAI-compatible)
	chatHandler := handler.NewChatHandler(g.modelRegistry, g.hooks)
	g.mux.HandleFunc("/v1/chat/completions", chatHandler.ServeHTTP)

	// Embeddings
	embeddingsHandler := handler.NewEmbeddingsHandler(g.modelRegistry, g.hooks)
	g.mux.HandleFunc("/v1/embeddings", embeddingsHandler.ServeHTTP)

	// Images
	imagesHandler := handler.NewImagesHandler(g.modelRegistry, g.hooks)
	g.mux.HandleFunc("/v1/images/generations", imagesHandler.ServeHTTP)

	// Health check
	g.mux.HandleFunc("/health", g.handleHealth)
	
	// Metrics endpoint (if metrics enabled)
	if g.metrics != nil {
		g.mux.Handle("/metrics", promhttp.Handler())
	}

	// 404 for unmatched routes
	g.mux.HandleFunc("/", g.handleNotFound)
}

// ServeHTTP implements http.Handler
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g.cors != nil {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if g.isOriginAllowed(origin) {
			// Set CORS headers
			if len(g.cors.AllowedOrigins) > 0 && g.cors.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				g.handlePreflight(w, r)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Set other CORS headers
			if g.cors.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(g.cors.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(g.cors.ExposedHeaders, ", "))
			}
		}
	}

	g.mux.ServeHTTP(w, r)
}

// isOriginAllowed checks if the origin is allowed
func (g *Gateway) isOriginAllowed(origin string) bool {
	if len(g.cors.AllowedOrigins) == 0 {
		return false
	}

	for _, allowed := range g.cors.AllowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}

	return false
}

// handlePreflight handles OPTIONS preflight requests
func (g *Gateway) handlePreflight(w http.ResponseWriter, r *http.Request) {
	// Set allowed methods
	if len(g.cors.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(g.cors.AllowedMethods, ", "))
	} else {
		// If no methods specified, allow the requested method
		if method := r.Header.Get("Access-Control-Request-Method"); method != "" {
			w.Header().Set("Access-Control-Allow-Methods", method)
		}
	}

	// Set allowed headers
	if len(g.cors.AllowedHeaders) > 0 {
		if g.cors.AllowedHeaders[0] == "*" {
			// Echo back the requested headers
			if headers := r.Header.Get("Access-Control-Request-Headers"); headers != "" {
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(g.cors.AllowedHeaders, ", "))
		}
	}

	// Set max age
	if g.cors.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", formatDuration(g.cors.MaxAge))
	}

	// Set allow credentials
	if g.cors.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// formatDuration converts a duration to seconds for Max-Age header
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.0f", d.Seconds())
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (g *Gateway) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":{"message":"Not found","type":"invalid_request_error"}}`))
}
