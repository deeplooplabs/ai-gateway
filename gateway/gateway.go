package gateway

import (
	"net/http"

	"github.com/deeplooplabs/ai-gateway/handler"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
)

// Gateway is the main HTTP handler
type Gateway struct {
	modelRegistry model.ModelRegistry
	hooks         *hook.Registry
	mux           *http.ServeMux
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
	// Chat Completions
	chatHandler := handler.NewChatHandler(g.modelRegistry, g.hooks)
	g.mux.HandleFunc("/v1/chat/completions", chatHandler.ServeHTTP)

	// Health check
	g.mux.HandleFunc("/health", g.handleHealth)

	// 404 for unmatched routes
	g.mux.HandleFunc("/", g.handleNotFound)
}

// ServeHTTP implements http.Handler
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
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
