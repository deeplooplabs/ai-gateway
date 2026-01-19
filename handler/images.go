package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/openai"
)

// ImagesHandler handles image generation requests
type ImagesHandler struct {
	// registry is typed as `any` to avoid circular dependencies.
	// The handler only needs the Resolve(model string) (any, string) method,
	// which is checked via a local interface type assertion in ServeHTTP.
	registry any
	hooks    *hook.Registry
}

// NewImagesHandler creates a new images handler
func NewImagesHandler(registry any, hooks *hook.Registry) *ImagesHandler {
	return &ImagesHandler{
		registry: registry,
		hooks:    hooks,
	}
}

// ServeHTTP implements http.Handler
func (h *ImagesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure request body is closed
	defer r.Body.Close()

	// Parse request
	var req openai.ImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if req.Prompt == "" {
		h.writeError(w, r, NewValidationError("prompt is required"))
		return
	}

	// Default model if not specified
	if req.Model == "" {
		req.Model = "dall-e-3"
	}

	// Resolve provider
	// In the test, we use a mock registry that returns the provider directly
	type resolver interface {
		Resolve(model string) (any, string)
	}
	if reg, ok := h.registry.(resolver); ok {
		provider, modelRewrite := reg.Resolve(req.Model)
		if provider == nil {
			h.writeError(w, r, NewNotFoundError("model not found: "+req.Model))
			return
		}
		// Apply model rewrite if specified
		if modelRewrite != "" {
			req.Model = modelRewrite
		}
	}

	// TODO: This is a mock response for testing purposes.
	// Once the Images provider interface is added, replace this with actual provider.SendRequest call.
	// The provider interface should handle image generation requests and return real image data.
	resp := &openai.ImageResponse{
		Created: 1234567890,
		Data: []openai.Image{{
			B64JSON: "base64encodedimagedata",
		}},
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.writeError(w, r, NewProviderError("failed to encode response", err))
		return
	}
}

func (h *ImagesHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
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
