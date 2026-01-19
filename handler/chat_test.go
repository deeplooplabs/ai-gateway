package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/openai"
	"github.com/deeplooplabs/ai-gateway/provider"
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

	var resp openai.ChatCompletionResponse
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

func newMockRegistry() *mapModelRegistry {
	prov := &mockChatProvider{}
	return &mapModelRegistry{provider: prov}
}

type mapModelRegistry struct {
	provider provider.Provider
}

func (m *mapModelRegistry) Resolve(model string) (provider.Provider, string) {
	return m.provider, ""
}

type mockChatProvider struct{}

func (m *mockChatProvider) Name() string {
	return "mock"
}

func (m *mockChatProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	return &openai.ChatCompletionResponse{
		ID:     "test-id",
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []openai.Choice{{
			Index: 0,
			Message: openai.Message{
				Role:    "assistant",
				Content: "Hello!",
			},
			FinishReason: "stop",
		}},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockChatProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
	chunkChan := make(chan openai.StreamChunk)
	errChan := make(chan error, 1)
	go func() {
		defer close(chunkChan)
		defer close(errChan)
	}()
	return chunkChan, errChan
}
