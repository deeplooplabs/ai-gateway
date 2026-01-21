package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	openai2 "github.com/deeplooplabs/ai-gateway/openresponses"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/provider"
)

// ResponsesHandler handles OpenResponses API requests
type ResponsesHandler struct {
	registry  model.ModelRegistry
	hooks     *hook.Registry
	orHooks   *openai2.Registry
	converter *openai2.Converter
}

// NewResponsesHandler creates a new responses handler
func NewResponsesHandler(registry model.ModelRegistry, hooks *hook.Registry) *ResponsesHandler {
	return &ResponsesHandler{
		registry:  registry,
		hooks:     hooks,
		orHooks:   openai2.NewRegistry(hooks),
		converter: openai2.NewConverter(),
	}
}

// SetOpenResponsesHooks sets OpenResponses-specific hooks
func (h *ResponsesHandler) SetOpenResponsesHooks(orHooks *openai2.Registry) {
	h.orHooks = orHooks
}

// ServeHTTP implements http.Handler for /v1/responses
func (h *ResponsesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Only POST is supported
	if r.Method != http.MethodPost {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"method_not_allowed",
			"Only POST method is allowed",
			"",
		))
		return
	}

	// Call AuthenticationHooks
	ctx := r.Context()
	for _, hh := range h.hooks.AuthenticationHooks() {
		success, tenantID, err := hh.Authenticate(ctx, r.Header.Get("Authorization"))
		if err != nil {
			h.writeError(w, r, openai2.NewError(
				"server_error",
				"authentication_failed",
				"Authentication failed: "+err.Error(),
				"",
			))
			return
		}
		if !success {
			h.writeError(w, r, openai2.NewError(
				"invalid_request_error",
				"unauthorized",
				"Invalid API key",
				"",
			))
			return
		}
		// Store tenantID in context
		ctx = context.WithValue(ctx, "tenant_id", tenantID)
		r = r.WithContext(ctx)
	}

	// Parse request
	var req openai2.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"invalid_json",
			"Invalid request body: "+err.Error(),
			"",
		))
		return
	}

	// Validate request
	if req.Model == "" {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"missing_model",
			"model is required",
			"model",
		))
		return
	}

	if req.Input == nil {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"missing_input",
			"input is required",
			"input",
		))
		return
	}

	// Set default truncation
	if req.Truncation == "" {
		req.Truncation = openai2.TruncationAuto
	}

	// Resolve provider
	prov, modelRewrite := h.registry.Resolve(req.Model)
	if prov == nil {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"model_not_found",
			fmt.Sprintf("Model not found: %s", req.Model),
			"model",
		))
		return
	}

	// Apply model rewrite if specified
	if modelRewrite != "" {
		req.Model = modelRewrite
	}

	// Handle streaming vs non-streaming
	stream := req.Stream != nil && *req.Stream
	if stream {
		h.handleStream(ctx, w, r, &req, prov)
		return
	}

	h.handleNonStream(ctx, w, r, &req, prov)
}

func (h *ResponsesHandler) handleNonStream(ctx context.Context, w http.ResponseWriter, r *http.Request, req *openai2.CreateRequest, prov provider.Provider) {
	// Generate response ID
	responseID := "resp_" + uuid.New().String()

	// Convert to OpenAI format
	chatReq, err := h.converter.RequestToChatCompletion(req)
	if err != nil {
		h.writeError(w, r, openai2.NewError(
			"invalid_request_error",
			"conversion_error",
			"Failed to convert request: "+err.Error(),
			"",
		))
		return
	}

	// Build unified request
	unifiedReq := provider.NewChatCompletionsRequest(chatReq.Model, chatReq.Messages)
	unifiedReq.Stream = false
	unifiedReq.Temperature = chatReq.Temperature
	unifiedReq.TopP = chatReq.TopP
	unifiedReq.MaxTokens = chatReq.MaxTokens
	unifiedReq.Stop = chatReq.Stop
	unifiedReq.PresencePenalty = chatReq.PresencePenalty
	unifiedReq.FrequencyPenalty = chatReq.FrequencyPenalty
	unifiedReq.Tools = chatReq.Tools
	unifiedReq.ToolChoice = chatReq.ToolChoice
	unifiedReq.Endpoint = "/v1/chat/completions"

	// Call BeforeRequest hooks
	for _, hh := range h.hooks.RequestHooks() {
		// Note: We're using OpenAI types for existing hooks
		// In the future, we'll have OpenResponses-specific hooks
		if err := hh.BeforeRequest(ctx, chatReq); err != nil {
			h.writeError(w, r, openai2.NewError(
				"server_error",
				"hook_error",
				"BeforeRequest hook failed: "+err.Error(),
				"",
			))
			return
		}
	}

	// Send request to provider using unified interface
	resp, err := prov.SendRequest(ctx, unifiedReq)
	if err != nil {
		h.writeError(w, r, openai2.NewError(
			"server_error",
			"provider_error",
			"Provider error: "+err.Error(),
			"",
		))
		return
	}
	defer resp.Close()

	// Convert response to OpenResponses format
	chatResp, err := resp.GetChatCompletion()
	if err != nil {
		h.writeError(w, r, openai2.NewError(
			"server_error",
			"conversion_error",
			"Failed to convert response: "+err.Error(),
			"",
		))
		return
	}
	if chatResp == nil {
		h.writeError(w, r, openai2.NewError(
			"server_error",
			"empty_response",
			"Empty response from provider",
			"",
		))
		return
	}

	orResp := h.converter.ChatCompletionToResponse(chatResp, responseID)

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orResp); err != nil {
		h.writeError(w, r, openai2.NewError(
			"server_error",
			"encode_error",
			"Failed to encode response: "+err.Error(),
			"",
		))
		return
	}
}

func (h *ResponsesHandler) handleStream(ctx context.Context, w http.ResponseWriter, r *http.Request, req *openai2.CreateRequest, prov provider.Provider) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, r, openai2.NewError(
			"server_error",
			"streaming_not_supported",
			"Streaming not supported",
			"",
		))
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create stream writer
	writer := openai2.NewStreamWriter(w, flusher)

	// Generate response ID
	responseID := "resp_" + uuid.New().String()

	// Send response.created event
	writer.WriteEvent(openai2.NewResponseCreatedEvent(writer.NextSequence(), responseID))

	// Send response.in_progress event
	writer.WriteEvent(openai2.NewResponseInProgressEvent(writer.NextSequence()))

	// Convert to OpenAI format
	chatReq, err := h.converter.RequestToChatCompletion(req)
	if err != nil {
		writer.WriteError(openai2.NewError(
			"invalid_request_error",
			"conversion_error",
			"Failed to convert request: "+err.Error(),
			"",
		))
		return
	}

	// Build unified request
	unifiedReq := provider.NewChatCompletionsRequest(chatReq.Model, chatReq.Messages)
	unifiedReq.Stream = true
	unifiedReq.Temperature = chatReq.Temperature
	unifiedReq.TopP = chatReq.TopP
	unifiedReq.MaxTokens = chatReq.MaxTokens
	unifiedReq.Stop = chatReq.Stop
	unifiedReq.PresencePenalty = chatReq.PresencePenalty
	unifiedReq.FrequencyPenalty = chatReq.FrequencyPenalty
	unifiedReq.Tools = chatReq.Tools
	unifiedReq.ToolChoice = chatReq.ToolChoice
	unifiedReq.Endpoint = "/v1/chat/completions"

	// Send request to provider using unified interface
	resp, err := prov.SendRequest(ctx, unifiedReq)
	if err != nil {
		writer.WriteError(openai2.NewError(
			"server_error",
			"provider_error",
			"Provider error: "+err.Error(),
			"",
		))
		return
	}
	defer resp.Close()

	if !resp.Stream {
		writer.WriteError(openai2.NewError(
			"server_error",
			"not_streaming",
			"Provider returned non-streaming response",
			"",
		))
		return
	}

	// Track state for item management
	seq := 0
	itemID := "msg_" + uuid.New().String()
	outputIndex := 0
	var itemAdded bool

	// Process chunks
	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-resp.Chunks:
			if !ok {
				// Channel closed, send completion
				orResp := openai2.NewResponse(responseID, req.Model)
				orResp.Status = openai2.ResponseStatusCompleted
				now := time.Now().Unix()
				orResp.CompletedAt = &now

				writer.WriteEvent(openai2.NewResponseCompletedEvent(writer.NextSequence(), orResp))
				writer.WriteDone()
				return
			}

			if chunk.Done {
				// Send completion
				orResp := openai2.NewResponse(responseID, req.Model)
				orResp.Status = openai2.ResponseStatusCompleted
				now := time.Now().Unix()
				orResp.CompletedAt = &now

				// Add completed message item if we haven't already
				if !itemAdded {
					messageItem := &openai2.MessageItem{
						ID:     itemID,
						Type:   "message",
						Status: openai2.MessageStatusCompleted,
						Role:   openai2.MessageRoleAssistant,
						Content: []openai2.OutputTextContent{
							{Type: "output_text", Text: ""},
						},
					}
					orResp.Output = []openai2.ItemField{messageItem}
				}

				writer.WriteEvent(openai2.NewResponseCompletedEvent(writer.NextSequence(), orResp))
				writer.WriteDone()
				return
			}

			// Process chunk based on type
			if chunk.Type == provider.ChunkTypeOpenAI && chunk.OpenAI != nil && len(chunk.OpenAI.Data) > 0 {
				// Send item added event if not sent yet
				if !itemAdded {
					messageItem := &openai2.MessageItem{
						ID:     itemID,
						Type:   "message",
						Status: openai2.MessageStatusInProgress,
						Role:   openai2.MessageRoleAssistant,
						Content: []openai2.OutputTextContent{
							{Type: "output_text", Text: ""},
						},
					}
					writer.WriteEvent(openai2.NewResponseOutputItemAddedEvent(writer.NextSequence(), outputIndex, messageItem))
					itemAdded = true

					// Send content part added event
					contentPart := openai2.OutputTextContent{Type: "output_text", Text: ""}
					writer.WriteEvent(openai2.NewResponseContentPartAddedEvent(writer.NextSequence(), itemID, outputIndex, 0, contentPart))
				}

				// Convert chunk to events
				events := h.converter.StreamingChunkToEvents(chunk.OpenAI.Data, &seq, itemID, outputIndex)

				// Apply streaming hooks and write events
				for _, event := range events {
					if err := writer.WriteEvent(event); err != nil {
						writer.WriteError(openai2.NewError(
							"server_error",
							"write_error",
							"Failed to write event: "+err.Error(),
							"",
						))
						return
					}
				}
			}

		case err := <-resp.Errors:
			if err != nil {
				writer.WriteError(openai2.NewError(
					"server_error",
					"stream_error",
					"Stream error: "+err.Error(),
					"",
				))
				return
			}
		}
	}
}

func (h *ResponsesHandler) writeError(w http.ResponseWriter, r *http.Request, err *openai2.Error) {
	// Call ErrorHooks
	ctx := r.Context()
	for _, hh := range h.hooks.ErrorHooks() {
		hh.OnError(ctx, fmt.Errorf("%s: %s", err.Type, err.Message))
	}

	// Set appropriate status code based on error type
	var statusCode int
	switch err.Type {
	case "invalid_request_error":
		statusCode = http.StatusBadRequest
	case "model_not_found":
		statusCode = http.StatusNotFound
	case "unauthorized":
		statusCode = http.StatusUnauthorized
	default:
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// OpenResponses error format
	errorResp := map[string]any{
		"error": err,
	}
	json.NewEncoder(w).Encode(errorResp)
}
