package provider

import (
	"context"
	"testing"

	"github.com/deeplooplabs/ai-gateway/openai"
)

func TestNewGeminiHTTPProvider(t *testing.T) {
	p := NewGeminiHTTPProvider("test-api-key")

	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Name() != "gemini-http" {
		t.Errorf("expected name 'gemini-http', got '%s'", p.Name())
	}
	if p.APIKey != "test-api-key" {
		t.Errorf("expected api key 'test-api-key', got '%s'", p.APIKey)
	}
}

func TestGeminiHTTPProvider_SendRequest(t *testing.T) {
	// This test verifies request construction without making real API calls
	p := NewGeminiHTTPProvider("test-key")

	// Test unsupported endpoint (images)
	ctx := context.Background()
	req := &openai.ChatCompletionRequest{Model: "gemini-pro"}

	_, err := p.SendRequest(ctx, "/v1/images/generations", req)
	if err == nil {
		t.Error("expected error for images endpoint, got nil")
	}

	// Verify the error message contains the expected text
	expectedErrMsg := "image generation not supported"
	if err != nil {
		errMsg := err.Error()
		if len(errMsg) < len(expectedErrMsg) || errMsg[:len(expectedErrMsg)] != expectedErrMsg {
			t.Errorf("expected error message starting with '%s', got '%s'", expectedErrMsg, errMsg)
		}
	}
}
