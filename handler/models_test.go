package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestModelsHandler_ServeHTTP_Success(t *testing.T) {
	// Setup mock registry with models
	registry := &mockModelsModelRegistry{
		models: []string{"gpt-4", "gpt-3.5-turbo", "text-embedding-3-small"},
	}

	handler := NewModelsHandler(registry)

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected 'application/json', got '%s'", contentType)
	}

	var resp openai.ModelsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected 'list', got '%s'", resp.Object)
	}

	expectedCount := 3
	if len(resp.Data) != expectedCount {
		t.Errorf("expected %d models, got %d", expectedCount, len(resp.Data))
	}

	// Verify model IDs
	modelIDs := make(map[string]bool)
	for _, model := range resp.Data {
		modelIDs[model.ID] = true
		if model.Object != "model" {
			t.Errorf("expected object 'model', got '%s'", model.Object)
		}
		if model.OwnedBy != "deeplooplabs" {
			t.Errorf("expected owned_by 'deeplooplabs', got '%s'", model.OwnedBy)
		}
	}

	expectedModels := []string{"gpt-4", "gpt-3.5-turbo", "text-embedding-3-small"}
	for _, modelID := range expectedModels {
		if !modelIDs[modelID] {
			t.Errorf("expected model '%s' not found in response", modelID)
		}
	}
}

func TestModelsHandler_ServeHTTP_EmptyList(t *testing.T) {
	// Setup mock registry with no models
	registry := &mockModelsModelRegistry{
		models: []string{},
	}

	handler := NewModelsHandler(registry)

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp openai.ModelsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("expected 0 models, got %d", len(resp.Data))
	}
}

func TestModelsHandler_ServeHTTP_PostMethodNotAllowed(t *testing.T) {
	registry := &mockModelsModelRegistry{
		models: []string{"gpt-4"},
	}

	handler := NewModelsHandler(registry)

	req := httptest.NewRequest("POST", "/v1/models", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

func TestModelsHandler_ServeHTTP_RegistryNotAvailable(t *testing.T) {
	// Pass something that doesn't implement the lister interface
	handler := NewModelsHandler("not a registry")

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

// mockModelsModelRegistry is a mock model registry for testing
type mockModelsModelRegistry struct {
	models []string
}

func (m *mockModelsModelRegistry) Resolve(model string) (provider.Provider, string) {
	return nil, ""
}

func (m *mockModelsModelRegistry) ResolveWithAPI(model string) (provider.Provider, string, provider.APIType) {
	return nil, "", 0
}

func (m *mockModelsModelRegistry) ListModels() []string {
	return m.models
}
