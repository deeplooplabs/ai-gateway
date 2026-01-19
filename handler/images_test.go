package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/hook"
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

type mockImagesProvider struct{}

func (m *mockImagesProvider) Name() string {
	return "mock-images"
}

func (m *mockImagesProvider) SendRequest(ctx context.Context, endpoint string, req any) (*openai.ImageResponse, error) {
	return &openai.ImageResponse{
		Created: 1234567890,
		Data: []openai.Image{{
			URL: "https://example.com/image.png",
		}},
	}, nil
}

type mockImagesRegistry struct {
	provider any
}

func (m *mockImagesRegistry) Resolve(model string) (any, string) {
	return m.provider, ""
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
