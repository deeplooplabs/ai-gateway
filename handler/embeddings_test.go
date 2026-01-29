package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
	openai2 "github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
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

	var resp openai.EmbeddingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected 'list', got '%s'", resp.Object)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(resp.Data))
	}
	if len(resp.Data[0].Embedding) != 3 {
		t.Errorf("expected 3 embedding values, got %d", len(resp.Data[0].Embedding))
	}
}

func TestEmbeddingsHandler_ServeHTTP_MultipleInputs(t *testing.T) {
	// Setup mock provider
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	// Create request with multiple inputs
	reqBody := map[string]any{
		"input": []string{"hello", "world"},
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

	var resp openai.EmbeddingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(resp.Data))
	}
}

func TestEmbeddingsHandler_ServeHTTP_EmptyModel(t *testing.T) {
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	reqBody := map[string]any{
		"input": "hello",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEmbeddingsHandler_ServeHTTP_EmptyInput(t *testing.T) {
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	reqBody := map[string]any{
		"model": "text-embedding-3-small",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEmbeddingsHandler_ServeHTTP_ModelNotFound(t *testing.T) {
	// Registry with no provider
	registry := &mockModelRegistry{provider: nil}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	reqBody := map[string]any{
		"input": "hello",
		"model": "unknown-model",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEmbeddingsHandler_ServeHTTP_WithDimensions(t *testing.T) {
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	reqBody := map[string]any{
		"input":      "hello world",
		"model":      "text-embedding-3-small",
		"dimensions": 512,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai.EmbeddingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(resp.Data))
	}
}

func TestEmbeddingsHandler_ServeHTTP_WithEncodingFormat(t *testing.T) {
	prov := &mockEmbeddingsProvider{}
	registry := &mockModelRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewEmbeddingsHandler(registry, hooks)

	reqBody := map[string]any{
		"input":           "hello world",
		"model":           "text-embedding-3-small",
		"encoding_format": "base64",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai.EmbeddingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(resp.Data))
	}
}

// mockEmbeddingsProvider is a mock provider that implements provider.Provider
type mockEmbeddingsProvider struct{}

func (m *mockEmbeddingsProvider) Name() string {
	return "mock-embeddings"
}

func (m *mockEmbeddingsProvider) SupportedAPIs() openai2.APIType {
	return openai2.APITypeEmbeddings
}

func (m *mockEmbeddingsProvider) SendRequest(ctx context.Context, req *openai2.Request) (*openai2.Response, error) {
	// Generate mock embedding response based on input
	inputCount := 1

	// Handle different input types (JSON unmarshaling can produce various types)
	switch v := req.EmbeddingInput.(type) {
	case []string:
		inputCount = len(v)
	case []interface{}:
		inputCount = len(v)
	case string:
		inputCount = 1
	}

	data := make([]openai.Embedding, inputCount)
	for i := range data {
		data[i] = openai.Embedding{
			Object:    "embedding",
			Embedding: []float32{0.1, 0.2, 0.3},
			Index:     i,
		}
	}

	resp := &openai.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  req.Model,
		Usage: openai.Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	return openai2.NewEmbeddingResponse(resp), nil
}

// mockModelRegistry is a mock model registry
type mockModelRegistry struct {
	provider openai2.Provider
}

func (m *mockModelRegistry) Resolve(model string) (openai2.Provider, string) {
	return m.provider, ""
}

func (m *mockModelRegistry) ResolveWithAPI(model string) (openai2.Provider, string, openai2.APIType) {
	return m.provider, "", openai2.APITypeEmbeddings
}

func (m *mockModelRegistry) ListModels() []string {
	return []string{"text-embedding-3-small"}
}
