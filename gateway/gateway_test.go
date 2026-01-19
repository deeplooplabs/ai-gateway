package gateway

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

func TestGateway_New(t *testing.T) {
	gw := New()

	if gw == nil {
		t.Error("expected non-nil gateway")
	}
}

func TestGateway_ServeHTTP_ChatCompletions(t *testing.T) {
	registry := setupTestRegistry()
	hooks := hook.NewRegistry()

	gw := New(
		WithModelRegistry(registry),
		WithHooks(hooks),
	)

	// Create chat completion request
	reqBody := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGateway_ServeHTTP_InvalidPath(t *testing.T) {
	gw := New()

	req := httptest.NewRequest("GET", "/invalid/path", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func setupTestRegistry() model.ModelRegistry {
	registry := model.NewMapModelRegistry()
	mockProv := &mockProvider{}
	registry.Register("gpt-4", mockProv, "")
	return registry
}

// mockProvider is a simple provider for testing
type mockProvider struct{}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) SendRequest(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (*openai2.ChatCompletionResponse, error) {
	return &openai2.ChatCompletionResponse{
		ID:     "test-id",
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []openai2.Choice{{
			Index: 0,
			Message: openai2.Message{
				Role:    "assistant",
				Content: "test response",
			},
			FinishReason: "stop",
		}},
	}, nil
}

func (m *mockProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (<-chan openai2.StreamChunk, <-chan error) {
	chunkChan := make(chan openai2.StreamChunk)
	errChan := make(chan error, 1)
	go func() {
		defer close(chunkChan)
		defer close(errChan)
	}()
	return chunkChan, errChan
}

// Ensure mockProvider implements the interface
var _ provider.Provider = (*mockProvider)(nil)

func TestGateway_EmbeddingsEndpoint(t *testing.T) {
	gw := New()

	reqBody := map[string]any{
		"input": "test",
		"model": "text-embedding-3-small",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("unexpected status: %d", w.Code)
	}
}

func TestGateway_ImagesEndpoint(t *testing.T) {
	gw := New()

	reqBody := map[string]any{
		"prompt": "a cat",
		"model":  "dall-e-3",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("unexpected status: %d", w.Code)
	}
}
