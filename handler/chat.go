package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/openai"
	"github.com/deeplooplabs/ai-gateway/provider"
)

// StreamingProvider is the provider interface for streaming requests
type StreamingProvider interface {
	provider.Provider
	SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
}

// ChatHandler handles chat completion requests
type ChatHandler struct {
	registry model.ModelRegistry
	hooks    *hook.Registry
}

// NewChatHandler creates a new chat handler
func NewChatHandler(registry model.ModelRegistry, hooks *hook.Registry) *ChatHandler {
	return &ChatHandler{
		registry: registry,
		hooks:    hooks,
	}
}

// ServeHTTP implements http.Handler
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure request body is closed
	defer r.Body.Close()

	// Call AuthenticationHooks to validate Authorization header
	for _, hh := range h.hooks.AuthenticationHooks() {
		success, userID, err := hh.Authenticate(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			h.writeError(w, r, fmt.Errorf("authentication failed: %w", err))
			return
		}
		if !success {
			h.writeError(w, r, NewValidationError("authentication failed"))
			return
		}
		// Store userID in request context for downstream use
		_ = userID // TODO: integrate userID into request context
	}

	// Parse request
	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if req.Model == "" {
		h.writeError(w, r, NewValidationError("model is required"))
		return
	}
	if len(req.Messages) == 0 {
		h.writeError(w, r, NewValidationError("messages is required"))
		return
	}

	// Resolve provider
	prov, modelRewrite := h.registry.Resolve(req.Model)
	if prov == nil {
		h.writeError(w, r, NewNotFoundError("model not found: "+req.Model))
		return
	}

	// Apply model rewrite if specified
	if modelRewrite != "" {
		req.Model = modelRewrite
	}

	// Handle streaming
	if req.Stream {
		h.handleStream(w, r, &req, prov)
		return
	}

	// Handle non-streaming
	h.handleNonStream(w, r, &req, prov)
}

func (h *ChatHandler) handleNonStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, prov provider.Provider) {
	// Call BeforeRequest hooks
	for _, hh := range h.hooks.RequestHooks() {
		if err := hh.BeforeRequest(r.Context(), req); err != nil {
			h.writeError(w, r, fmt.Errorf("hook error: %w", err))
			return
		}
	}

	// Send request to provider
	resp, err := prov.SendRequest(r.Context(), "/v1/chat/completions", req)
	if err != nil {
		h.writeError(w, r, NewProviderError("provider error", err))
		return
	}

	// Call AfterRequest hooks
	for _, hh := range h.hooks.RequestHooks() {
		if err := hh.AfterRequest(r.Context(), req, resp); err != nil {
			h.writeError(w, r, fmt.Errorf("hook error: %w", err))
			return
		}
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.writeError(w, r, NewProviderError("failed to encode response", err))
		return
	}
}

func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, prov provider.Provider) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, r, NewValidationError("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Type assert to StreamingProvider
	streamingProvider, ok := prov.(StreamingProvider)
	if !ok {
		h.writeError(w, r, NewValidationError("provider does not support streaming"))
		return
	}

	// Get streaming channels from provider
	chunkChan, errChan := streamingProvider.SendRequestStream(r.Context(), "/v1/chat/completions", req)

	// Process chunks
	for {
		select {
		case <-r.Context().Done():
			return
		case chunk, ok := <-chunkChan:
			if !ok {
				return
			}
			if chunk.Done {
				// Send [DONE] marker
				io.WriteString(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			if len(chunk.Data) > 0 {
				// Call streaming hooks
				modifiedData := chunk.Data
				for _, hh := range h.hooks.StreamingHooks() {
					result, err := hh.OnChunk(r.Context(), modifiedData)
					if err != nil {
						h.writeError(w, r, fmt.Errorf("streaming hook error: %w", err))
						return
					}
					modifiedData = result
				}

				// Write SSE formatted chunk
				io.WriteString(w, "data: ")
				io.WriteString(w, string(modifiedData))
				io.WriteString(w, "\n\n")
				flusher.Flush()
			}
		case err := <-errChan:
			if err != nil {
				h.writeError(w, r, NewProviderError("stream error", err))
				return
			}
		}
	}
}

func (h *ChatHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
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

// GatewayError represents a gateway error (simplified for handler)
type GatewayError struct {
	Code    int
	Message string
	Type    string
	Err     error
}

func NewValidationError(msg string) *GatewayError {
	return &GatewayError{Code: 400, Message: msg, Type: "invalid_request_error"}
}

func NewNotFoundError(msg string) *GatewayError {
	return &GatewayError{Code: 404, Message: msg, Type: "not_found_error"}
}

func NewProviderError(msg string, err error) *GatewayError {
	return &GatewayError{Code: 502, Message: msg, Type: "api_error", Err: err}
}

func (e *GatewayError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *GatewayError) ToOpenAIResponse() map[string]any {
	return map[string]any{
		"error": map[string]any{
			"message": e.Message,
			"type":    e.Type,
		},
	}
}
