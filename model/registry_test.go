package model

import (
	"context"
	"testing"

	"github.com/deeplooplabs/ai-gateway/provider"
)

func TestMapModelRegistry(t *testing.T) {
	// Use provider package to avoid import error
	var _ provider.Provider = (*mockProvider)(nil)

	prov1 := &mockProvider{name: "provider1"}
	prov2 := &mockProvider{name: "provider2"}

	registry := NewMapModelRegistry()

	// Register models using new API
	registry.Register("gpt-4", prov1)
	registry.RegisterWithOptions("gpt-3.5-turbo", prov2, WithModelRewrite("gpt-35-turbo"))
	registry.Register("claude-3", prov1)

	// Test exact match
	p, modelRewrite := registry.Resolve("gpt-4")
	if p.Name() != "provider1" {
		t.Errorf("expected 'provider1', got '%s'", p.Name())
	}
	if modelRewrite != "" {
		t.Errorf("expected empty rewrite, got '%s'", modelRewrite)
	}

	// Test model rewrite
	p, modelRewrite = registry.Resolve("gpt-3.5-turbo")
	if p.Name() != "provider2" {
		t.Errorf("expected 'provider2', got '%s'", p.Name())
	}
	if modelRewrite != "gpt-35-turbo" {
		t.Errorf("expected 'gpt-35-turbo', got '%s'", modelRewrite)
	}

	// Test unknown model (should still work, returns nil provider)
	p, modelRewrite = registry.Resolve("unknown")
	if p != nil {
		t.Error("expected nil provider for unknown model")
	}
}

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) SupportedAPIs() provider.APIType {
	return provider.APITypeChatCompletions
}

func (m *mockProvider) SendRequest(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	return nil, nil
}

func TestMapModelRegistry_GeminiProvider(t *testing.T) {
	registry := NewMapModelRegistry()

	mockProvider := &mockProvider{name: "test-gemini"}
	registry.RegisterWithOptions("gemini-pro", mockProvider, WithModelRewrite("gemini-2.0-flash-exp"))

	prov, rewrite := registry.Resolve("gemini-pro")

	if prov == nil {
		t.Fatal("expected non-nil provider")
	}
	if prov.Name() != "test-gemini" {
		t.Errorf("expected provider name 'test-gemini', got '%s'", prov.Name())
	}
	if rewrite != "gemini-2.0-flash-exp" {
		t.Errorf("expected rewrite 'gemini-2.0-flash-exp', got '%s'", rewrite)
	}
}

func TestMapModelRegistry_ResolveWithAPI(t *testing.T) {
	registry := NewMapModelRegistry()

	prov1 := &mockProvider{name: "provider1"}

	// Register with preferred API type
	registry.RegisterWithOptions("gpt-4", prov1, WithPreferredAPI(provider.APITypeResponses))

	// Test ResolveWithAPI
	p, _, apiType := registry.ResolveWithAPI("gpt-4")

	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Name() != "provider1" {
		t.Errorf("expected provider name 'provider1', got '%s'", p.Name())
	}
	if apiType != provider.APITypeResponses {
		t.Errorf("expected APITypeResponses, got '%v'", apiType)
	}
}
