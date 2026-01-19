package provider

import (
	"testing"
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
