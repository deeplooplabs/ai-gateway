package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// ModelsHandler handles model list requests
type ModelsHandler struct {
	// registry is typed as `any` to avoid circular dependencies.
	// The handler only needs the ListModels() []string method,
	// which is checked via a local interface type assertion in ServeHTTP.
	registry any
}

// NewModelsHandler creates a new models handler
func NewModelsHandler(registry any) *ModelsHandler {
	return &ModelsHandler{
		registry: registry,
	}
}

// ServeHTTP implements http.Handler
func (h *ModelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only GET method is supported
	if r.Method != http.MethodGet {
		h.writeError(w, NewMethodNotAllowedError("only GET method is allowed"))
		return
	}

	// Get model list from registry
	type lister interface {
		ListModels() []string
	}

	var models []string
	if reg, ok := h.registry.(lister); ok {
		models = reg.ListModels()
	} else {
		h.writeError(w, NewProviderError("registry not available", nil))
		return
	}

	// Sort models for consistent output
	sort.Strings(models)

	// Build response
	now := time.Now().Unix()
	modelData := make([]openai.Model, 0, len(models))
	for _, modelID := range models {
		modelData = append(modelData, openai.Model{
			ID:      modelID,
			Object:  "model",
			Created: now,
			OwnedBy: "deeplooplabs",
		})
	}

	resp := openai.ModelsResponse{
		Object: "list",
		Data:   modelData,
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.writeError(w, NewProviderError("failed to encode response", err))
		return
	}
}

func (h *ModelsHandler) writeError(w http.ResponseWriter, err error) {
	var gwErr *GatewayError
	if e, ok := err.(*GatewayError); ok {
		gwErr = e
	} else {
		gwErr = NewProviderError("internal error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(gwErr.Code)
	json.NewEncoder(w).Encode(gwErr.ToOpenAIResponse())
}
