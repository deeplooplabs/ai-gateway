package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// EmbeddingsHandler handles embedding requests
type EmbeddingsHandler struct {
	// registry is typed as `any` to avoid circular dependencies.
	// The handler only needs the Resolve(model string) (provider.Provider, string) method,
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

	ctx := r.Context()

	// Log incoming request
	slog.InfoContext(ctx, "Embeddings request received",
		"model", req.Model,
		"encoding_format", req.EncodingFormat,
		"dimensions", req.Dimensions,
	)

	// Resolve provider
	type resolver interface {
		Resolve(model string) (provider.Provider, string)
	}
	var prov provider.Provider
	var modelRewrite string

	if reg, ok := h.registry.(resolver); ok {
		prov, modelRewrite = reg.Resolve(req.Model)
		if prov == nil {
			h.writeError(w, r, NewNotFoundError("model not found: "+req.Model))
			return
		}
	} else {
		h.writeError(w, r, NewProviderError("registry not available", nil))
		return
	}

	// Apply model rewrite if specified
	if modelRewrite != "" {
		slog.InfoContext(ctx, "Model rewrite applied",
			"original", req.Model,
			"rewritten", modelRewrite,
		)
		req.Model = modelRewrite
	}

	slog.InfoContext(ctx, "Provider resolved",
		"provider", prov.Name(),
		"supported_apis", prov.SupportedAPIs().String(),
	)

	// Create provider request
	provReq := provider.NewEmbeddingsRequest(req.Model, req.Input)
	provReq.EncodingFormat = req.EncodingFormat
	provReq.Dimensions = req.Dimensions

	// Send request to provider
	provResp, err := prov.SendRequest(ctx, provReq)
	if err != nil {
		slog.ErrorContext(ctx, "Provider request failed",
			"error", err.Error(),
			"provider", prov.Name(),
			"model", req.Model,
		)
		h.writeError(w, r, NewProviderError("provider request failed", err))
		return
	}

	// Get embedding response
	resp, err := provResp.GetEmbedding()
	if err != nil {
		h.writeError(w, r, NewProviderError("invalid response", err))
		return
	}

	slog.InfoContext(ctx, "Embeddings response successful",
		"embedding_count", len(resp.Data),
		"model", resp.Model,
		"prompt_tokens", resp.Usage.PromptTokens,
	)

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
	if h.hooks != nil {
		for _, hh := range h.hooks.ErrorHooks() {
			hh.OnError(ctx, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(gwErr.Code)
	if encodeErr := json.NewEncoder(w).Encode(gwErr.ToOpenAIResponse()); encodeErr != nil {
		fmt.Printf("failed to encode error response: %v\n", encodeErr)
	}
}
