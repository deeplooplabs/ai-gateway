package model

import (
	"context"
	"testing"

	"github.com/deeplooplabs/ai-gateway/openai"
	"github.com/deeplooplabs/ai-gateway/provider"
)

func TestMapModelRegistry(t *testing.T) {
	// Use provider package to avoid import error
	var _ provider.Provider = (*mockProvider)(nil)

	prov1 := &mockProvider{name: "provider1"}
	prov2 := &mockProvider{name: "provider2"}

	registry := NewMapModelRegistry()

	// Register models
	registry.Register("gpt-4", prov1, "")
	registry.Register("gpt-3.5-turbo", prov2, "gpt-35-turbo")
	registry.Register("claude-3", prov1, "")

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

func (m *mockProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	return nil, nil
}
