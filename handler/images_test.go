package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
	prov "github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestImagesHandler_ServeHTTP(t *testing.T) {
	// Setup mock provider
	prov := &mockImagesProvider{}
	registry := &mockImagesRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewImagesHandler(registry, hooks)

	// Create request
	reqBody := map[string]any{
		"model":  "dall-e-3",
		"prompt": "a cat",
		"n":      1,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai.ImageResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 image, got %d", len(resp.Data))
	}
}

func TestImagesHandler_MissingPrompt(t *testing.T) {
	registry := &mockImagesRegistry{provider: &mockImagesProvider{}}
	hooks := hook.NewRegistry()

	handler := NewImagesHandler(registry, hooks)

	// Create request without prompt
	reqBody := map[string]any{
		"model": "dall-e-3",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["error"] == nil {
		t.Errorf("expected error response, got %#v", resp)
	}
}

func TestImagesHandler_DefaultModel(t *testing.T) {
	prov := &mockImagesProvider{}
	registry := &mockImagesRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewImagesHandler(registry, hooks)

	// Create request without model
	reqBody := map[string]any{
		"prompt": "a cat",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestImagesHandler_ModelNotFound(t *testing.T) {
	// Registry with no provider
	registry := &mockImagesRegistry{provider: nil}
	hooks := hook.NewRegistry()

	handler := NewImagesHandler(registry, hooks)

	reqBody := map[string]any{
		"prompt": "a cat",
		"model":  "unknown-model",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestImagesHandler_WithOptions(t *testing.T) {
	prov := &mockImagesProvider{}
	registry := &mockImagesRegistry{provider: prov}
	hooks := hook.NewRegistry()

	handler := NewImagesHandler(registry, hooks)

	reqBody := map[string]any{
		"prompt":  "a cat",
		"model":   "dall-e-3",
		"n":       2,
		"size":    "1024x1024",
		"quality": "hd",
		"style":   "vivid",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai.ImageResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 images, got %d", len(resp.Data))
	}
}

// mockImagesProvider is a mock provider that implements provider.Provider
type mockImagesProvider struct{}

func (m *mockImagesProvider) Name() string {
	return "mock-images"
}

func (m *mockImagesProvider) SupportedAPIs() prov.APIType {
	return prov.APITypeImages
}

func (m *mockImagesProvider) SendRequest(ctx context.Context, req *prov.Request) (*prov.Response, error) {
	// Get N from request (default to 1)
	n := 1
	if req.ImageN > 0 {
		n = req.ImageN
	}

	data := make([]openai.Image, n)
	for i := range data {
		data[i] = openai.Image{
			URL: "https://example.com/image.png",
		}
	}

	resp := &openai.ImageResponse{
		Created: 1234567890,
		Data:    data,
	}

	return prov.NewImageResponse(resp), nil
}

// mockImagesRegistry is a mock model registry
type mockImagesRegistry struct {
	provider prov.Provider
}

func (m *mockImagesRegistry) Resolve(model string) (prov.Provider, string) {
	return m.provider, ""
}
