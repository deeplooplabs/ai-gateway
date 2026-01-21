package provider

import (
	"context"
)

// Provider defines the unified interface for LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// SupportedAPIs returns the API type(s) this provider supports
	SupportedAPIs() APIType

	// SendRequest sends a request (streaming or non-streaming based on req.Stream)
	// Returns a Response that can be either a complete response or a streaming channel
	SendRequest(ctx context.Context, req *Request) (*Response, error)
}

// Ensure BaseProvider implements Provider
var _ Provider = (*BaseProvider)(nil)
