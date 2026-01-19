package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestEmbeddingsHandler_ServeHTTP(t *testing.T) {
	// Setup mock provider
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	// Create request
	reqBody := map[string]any{
		"input": "hello world",
		"model": "text-embedding-3-small",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai2.EmbeddingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected 'list', got '%s'", resp.Object)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(resp.Data))
	}
}

type mockEmbeddingsProvider struct{}

func (m *mockEmbeddingsProvider) Name() string {
	return "mock-embeddings"
}

func (m *mockEmbeddingsProvider) SendRequest(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (*openai2.ChatCompletionResponse, error) {
	return &openai2.ChatCompletionResponse{}, nil
}

func (m *mockEmbeddingsProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (<-chan openai2.StreamChunk, <-chan error) {
	return nil, nil
}

type mockModelRegistry struct {
	provider any
}

func (m *mockModelRegistry) Resolve(model string) (any, string) {
	return m.provider, ""
}
