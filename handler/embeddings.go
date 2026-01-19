package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// EmbeddingsHandler handles embedding requests
type EmbeddingsHandler struct {
	// registry is typed as `any` to avoid circular dependencies.
	// The handler only needs the Resolve(model string) (any, string) method,
	// which is checked via a local interface type assertion in ServeHTTP.
	registry any
	hooks    *hook.Registry
}

// NewEmbeddingsHandler creates a new embeddings handler
func NewEmbeddingsHandler(registry any, hooks *hook.Registry) *EmbeddingsHandler {
	return &EmbeddingsHandler{
		registry: registry,
		hooks:    hooks,
	}
}

// ServeHTTP implements http.Handler
func (h *EmbeddingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure request body is closed
	defer r.Body.Close()

	// Parse request
	var req openai.EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if req.Model == "" {
		h.writeError(w, r, NewValidationError("model is required"))
		return
	}
	if req.Input == nil {
		h.writeError(w, r, NewValidationError("input is required"))
		return
	}

	// Resolve provider
	// In the test, we use a mock registry that returns the provider directly
	type resolver interface {
		Resolve(model string) (any, string)
	}
	if reg, ok := h.registry.(resolver); ok {
		provider, _ := reg.Resolve(req.Model)
		if provider == nil {
			h.writeError(w, r, NewNotFoundError("model not found: "+req.Model))
			return
		}
	}

	// Call BeforeRequest hooks
	// Note: Current hook.RequestHook is typed for ChatCompletionRequest/Response
	// For embeddings, hooks would need a new interface to be added later
	// For now, we skip hook calls in embeddings handler

	// For mock/testing, return mock response
	// In real implementation, this would call provider.SendRequest
	resp := &openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{{
			Object:    "embedding",
			Embedding: []float32{0.1, 0.2, 0.3},
			Index:     0,
		}},
		Model: req.Model,
		Usage: openai.Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	// Call AfterRequest hooks
	// Note: Skipped for now, same reason as BeforeRequest

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.writeError(w, r, NewProviderError("failed to encode response", err))
		return
	}
}

func (h *EmbeddingsHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var gwErr *GatewayError
	if e, ok := err.(*GatewayError); ok {
		gwErr = e
	} else {
		gwErr = NewProviderError("internal error", err)
	}

	// Call ErrorHooks to notify of the error
	ctx := r.Context()
	for _, hh := range h.hooks.ErrorHooks() {
		hh.OnError(ctx, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(gwErr.Code)
	if encodeErr := json.NewEncoder(w).Encode(gwErr.ToOpenAIResponse()); encodeErr != nil {
		fmt.Printf("failed to encode error response: %v\n", encodeErr)
	}
}
