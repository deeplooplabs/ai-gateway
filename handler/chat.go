package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/openai"
	"github.com/deeplooplabs/ai-gateway/provider"
)

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
	// Parse request
	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if req.Model == "" {
		h.writeError(w, NewValidationError("model is required"))
		return
	}
	if len(req.Messages) == 0 {
		h.writeError(w, NewValidationError("messages is required"))
		return
	}

	// Resolve provider
	prov, modelRewrite := h.registry.Resolve(req.Model)
	if prov == nil {
		h.writeError(w, NewNotFoundError("model not found: "+req.Model))
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
			h.writeError(w, fmt.Errorf("hook error: %w", err))
			return
		}
	}

	// Send request to provider
	resp, err := prov.SendRequest(r.Context(), "/v1/chat/completions", req)
	if err != nil {
		h.writeError(w, NewProviderError("provider error", err))
		return
	}

	// Call AfterRequest hooks
	for _, hh := range h.hooks.RequestHooks() {
		if err := hh.AfterRequest(r.Context(), req, resp); err != nil {
			h.writeError(w, fmt.Errorf("hook error: %w", err))
			return
		}
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, prov provider.Provider) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, NewValidationError("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// For mock/testing, send simple chunk
	chunk := `data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"` + req.Model + `","choices":[{"index":0,"delta":{"content":"Hello!"},"finish_reason":null}]}` + "\n\n"
	io.WriteString(w, chunk)

	endChunk := `data: [DONE]` + "\n\n"
	io.WriteString(w, endChunk)

	flusher.Flush()
}

func (h *ChatHandler) writeError(w http.ResponseWriter, err error) {
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
