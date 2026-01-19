package ai_gateway

import (
	"fmt"
	"net/http"
)

// GatewayError represents an error that can be returned to the client
type GatewayError struct {
	Code       int
	Message    string
	Type       string
	Param      string
	InnerError error
}

// Error implements the error interface
func (e *GatewayError) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("%s: %s (inner: %v)", e.Type, e.Message, e.InnerError)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// OpenAIErrorResponse represents the error response format compatible with OpenAI
type OpenAIErrorResponse struct {
	Error *OpenAIErrorDetail `json:"error"`
}

// OpenAIErrorDetail represents the error detail in OpenAI format
type OpenAIErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ToOpenAIResponse converts the GatewayError to OpenAI error response format
func (e *GatewayError) ToOpenAIResponse() *OpenAIErrorResponse {
	return &OpenAIErrorResponse{
		Error: &OpenAIErrorDetail{
			Message: e.Message,
			Type:    e.Type,
			Param:   e.Param,
			Code:    http.StatusText(e.Code),
		},
	}
}

// NewAuthenticationError creates a new authentication error (401)
func NewAuthenticationError(message string) *GatewayError {
	return &GatewayError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Type:    "authentication_error",
	}
}

// NewValidationError creates a new validation error (400)
func NewValidationError(message string) *GatewayError {
	return &GatewayError{
		Code:    http.StatusBadRequest,
		Message: message,
		Type:    "invalid_request_error",
	}
}

// NewNotFoundError creates a new not found error (404)
func NewNotFoundError(message string) *GatewayError {
	return &GatewayError{
		Code:    http.StatusNotFound,
		Message: message,
		Type:    "invalid_request_error",
	}
}

// NewRateLimitError creates a new rate limit error (429)
func NewRateLimitError(message string) *GatewayError {
	return &GatewayError{
		Code:    http.StatusTooManyRequests,
		Message: message,
		Type:    "rate_limit_error",
	}
}

// NewProviderError creates a new provider error (502)
func NewProviderError(message string, inner error) *GatewayError {
	return &GatewayError{
		Code:       http.StatusBadGateway,
		Message:    message,
		Type:       "api_error",
		InnerError: inner,
	}
}
