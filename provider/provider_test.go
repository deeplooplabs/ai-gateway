package provider

import (
	"context"
	"testing"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestProviderInterface(t *testing.T) {
	// Mock provider for testing
	var p Provider = &mockProvider{}

	if p.Name() != "mock" {
		t.Errorf("expected 'mock', got '%s'", p.Name())
	}

	if p.SupportedAPIs() != APITypeChatCompletions {
		t.Errorf("expected APITypeChatCompletions, got '%v'", p.SupportedAPIs())
	}

	req := NewChatCompletionsRequest("gpt-4", []openai2.Message{{Role: "user", Content: "test"}})

	resp, err := p.SendRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
	}
	if resp.Stream {
		t.Error("expected non-streaming response")
	}
}

type mockProvider struct{}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) SupportedAPIs() APIType {
	return APITypeChatCompletions
}

func (m *mockProvider) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	// Return a mock Chat Completions response
	chatResp := &openai2.ChatCompletionResponse{
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
	}
	return NewChatCompletionResponse(chatResp), nil
}

func TestHTTPProvider(t *testing.T) {
	config := NewProviderConfig("http").
		WithBaseURL("https://api.openai.com").
		WithAPIKey("test-key")

	provider := NewHTTPProvider(config)

	if provider.Name() != "http" {
		t.Errorf("expected 'http', got '%s'", provider.Name())
	}

	if provider.Config().BaseURL != "https://api.openai.com" {
		t.Errorf("expected 'https://api.openai.com', got '%s'", provider.Config().BaseURL)
	}
}

func TestHTTPProviderWithBaseURL(t *testing.T) {
	provider := NewHTTPProviderWithBaseURL("https://api.openai.com", "test-key")

	if provider.Name() != "http" {
		t.Errorf("expected 'http', got '%s'", provider.Name())
	}

	if provider.Config().BaseURL != "https://api.openai.com" {
		t.Errorf("expected 'https://api.openai.com', got '%s'", provider.Config().BaseURL)
	}
}

func TestProviderStreamingInterface(t *testing.T) {
	// Test that mock provider implements streaming
	var p Provider = &mockStreamProvider{}

	if p.SupportedAPIs() != APITypeAll {
		t.Errorf("expected APITypeAll, got '%v'", p.SupportedAPIs())
	}

	// Non-streaming should work
	req := NewChatCompletionsRequest("gpt-4", []openai2.Message{{Role: "user", Content: "test"}})
	req.Stream = true
	resp, err := p.SendRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
	}
	if !resp.Stream {
		t.Error("expected streaming response")
	}

	// Receive chunks
	chunkCount := 0
	for chunk := range resp.Chunks {
		chunkCount++
		if chunk.Done {
			break
		}
	}

	if chunkCount != 3 {
		t.Errorf("expected 3 chunks, got %d", chunkCount)
	}

	resp.Close()
}

type mockStreamProvider struct{}

func (m *mockStreamProvider) Name() string {
	return "mock-stream"
}

func (m *mockStreamProvider) SupportedAPIs() APIType {
	return APITypeAll
}

func (m *mockStreamProvider) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	if req.Stream {
		// Return a streaming response
		chunkChan := make(chan *Chunk, 2)
		errChan := make(chan error, 1)

		go func() {
			defer close(chunkChan)
			defer close(errChan)

			// Send test chunks
			chunkChan <- NewOpenAIChunk([]byte(`{"id":"test","choices":[{"delta":{"content":"Hello"}}]}`))
			chunkChan <- NewOpenAIChunk([]byte(`{"id":"test","choices":[{"delta":{"content":" world"}}]}`))
			chunkChan <- NewOpenAIChunkDone()
		}()

		return NewStreamingResponse(APITypeChatCompletions, chunkChan, errChan, func() error { return nil }), nil
	}

	// Non-streaming response
	chatResp := &openai2.ChatCompletionResponse{
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
	}
	return NewChatCompletionResponse(chatResp), nil
}
