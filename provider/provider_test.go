package provider

import (
	"context"
	"testing"

	"github.com/deeplooplabs/ai-gateway/openai"
)

func TestProviderInterface(t *testing.T) {
	// Mock provider for testing
	var p Provider = &mockProvider{
		baseURL: "https://api.openai.com",
	}

	if p.Name() != "mock" {
		t.Errorf("expected 'mock', got '%s'", p.Name())
	}

	req := &openai.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []openai.Message{{Role: "user", Content: "test"}},
	}

	resp, err := p.SendRequest(context.Background(), "/v1/chat/completions", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
	}
}

type mockProvider struct {
	baseURL string
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	return &openai.ChatCompletionResponse{
		ID:     "test-id",
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []openai.Choice{{
			Index: 0,
			Message: openai.Message{
				Role:    "assistant",
				Content: "test response",
			},
			FinishReason: "stop",
		}},
	}, nil
}

func TestHTTPProvider(t *testing.T) {
	provider := NewHTTPProvider("https://api.openai.com", "test-key")

	if provider.Name() != "http" {
		t.Errorf("expected 'http', got '%s'", provider.Name())
	}

	if provider.BaseURL != "https://api.openai.com" {
		t.Errorf("expected 'https://api.openai.com', got '%s'", provider.BaseURL)
	}
}
