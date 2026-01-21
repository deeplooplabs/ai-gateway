package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/provider"
	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestChatHandler_ServeHTTP(t *testing.T) {
	// Setup
	registry := newMockRegistry()
	hooks := hook.NewRegistry()

	handler := NewChatHandler(registry, hooks)

	// Create request
	reqBody := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(w, req)

	// Verify
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai2.ChatCompletionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Object != "chat.completion" {
		t.Errorf("expected 'chat.completion', got '%s'", resp.Object)
	}
}

func TestChatHandler_Stream(t *testing.T) {
	registry := newMockRegistry()
	hooks := hook.NewRegistry()

	handler := NewChatHandler(registry, hooks)

	reqBody := map[string]any{
		"model":    "gpt-4",
		"messages": []map[string]string{{"role": "user", "content": "Hello"}},
		"stream":   true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected 'text/event-stream', got '%s'", contentType)
	}
}

func newMockRegistry() model.ModelRegistry {
	prov := &mockChatProvider{}
	return &mapModelRegistry{provider: prov}
}

type mapModelRegistry struct {
	provider provider.Provider
}

func (m *mapModelRegistry) Resolve(model string) (provider.Provider, string) {
	return m.provider, ""
}

func (m *mapModelRegistry) ResolveWithAPI(model string) (provider.Provider, string, provider.APIType) {
	return m.provider, "", provider.APITypeChatCompletions
}

type mockChatProvider struct{}

func (m *mockChatProvider) Name() string {
	return "mock"
}

func (m *mockChatProvider) SupportedAPIs() provider.APIType {
	return provider.APITypeChatCompletions
}

func (m *mockChatProvider) SendRequest(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if req.Stream {
		// Return streaming response
		chunkChan := make(chan *provider.Chunk, 2)
		errChan := make(chan error, 1)

		go func() {
			defer close(chunkChan)
			defer close(errChan)

			// Send a mock chunk
			chunkData := `{"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"` + req.Model + `","choices":[{"index":0,"delta":{"content":"Hello!"},"finish_reason":null}]}`
			chunkChan <- provider.NewOpenAIChunk([]byte(chunkData))

			// Send done marker
			chunkChan <- provider.NewOpenAIChunkDone()
		}()

		return provider.NewStreamingResponse(provider.APITypeChatCompletions, chunkChan, errChan, func() error { return nil }), nil
	}

	// Return non-streaming response
	return provider.NewChatCompletionResponse(&openai2.ChatCompletionResponse{
		ID:     "test-id",
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []openai2.Choice{{
			Index: 0,
			Message: openai2.Message{
				Role:    "assistant",
				Content: "Hello!",
			},
			FinishReason: "stop",
		}},
		Usage: openai2.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}), nil
}
