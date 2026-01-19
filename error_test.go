package ai_gateway

import (
	"net/http"
	"testing"
)

func TestGatewayError(t *testing.T) {
	err := &GatewayError{
		Code:    http.StatusBadRequest,
		Message: "Invalid API key",
		Type:    "invalid_request_error",
	}

	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}

	// Test JSON marshaling to OpenAI format
	resp := err.ToOpenAIResponse()
	if resp.Error == nil {
		t.Error("ToOpenAIResponse should have Error field")
	}
	if resp.Error.Message != "Invalid API key" {
		t.Errorf("expected 'Invalid API key', got '%s'", resp.Error.Message)
	}
	if resp.Error.Type != "invalid_request_error" {
		t.Errorf("expected 'invalid_request_error', got '%s'", resp.Error.Type)
	}
}

func TestNewAuthenticationError(t *testing.T) {
	err := NewAuthenticationError("Invalid API key")
	if err.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", err.Code)
	}
	if err.Type != "authentication_error" {
		t.Errorf("expected 'authentication_error', got '%s'", err.Type)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("model is required")
	if err.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.Code)
	}
	if err.Type != "invalid_request_error" {
		t.Errorf("expected 'invalid_request_error', got '%s'", err.Type)
	}
}
