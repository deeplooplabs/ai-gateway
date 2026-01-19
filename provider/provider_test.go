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

func (m *mockProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
	chunkChan := make(chan openai.StreamChunk, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)
		chunkChan <- openai.StreamChunk{Done: true}
	}()

	return chunkChan, errChan
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

func TestProviderStreamingInterface(t *testing.T) {
	// Test that mock provider implements streaming
	var p Provider = &mockStreamProvider{}

	// Non-streaming should work
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

	// Streaming should also work through the interface
	chunkChan, errChan := p.SendRequestStream(context.Background(), "/v1/chat/completions", req)

	// Receive chunks
	chunkCount := 0
	for chunk := range chunkChan {
		chunkCount++
		if chunk.Done {
			break
		}
	}

	// Check for errors
	for range errChan {
		// No errors expected
	}

	if chunkCount != 3 {
		t.Errorf("expected 3 chunks, got %d", chunkCount)
	}
}

type mockStreamProvider struct{}

func (m *mockStreamProvider) Name() string {
	return "mock-stream"
}

func (m *mockStreamProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
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

func (m *mockStreamProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
	chunkChan := make(chan openai.StreamChunk, 2)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		// Send test chunks
		chunkChan <- openai.StreamChunk{Data: []byte(`{"id":"test","choices":[{"delta":{"content":"Hello"}}]}`)}
		chunkChan <- openai.StreamChunk{Data: []byte(`{"id":"test","choices":[{"delta":{"content":" world"}}]}`)}
		chunkChan <- openai.StreamChunk{Done: true}
	}()

	return chunkChan, errChan
}
