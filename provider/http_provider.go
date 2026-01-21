package provider

import (
	"context"
	"fmt"
)

// HTTPProvider is a configurable HTTP-based provider that supports both Chat Completions and Responses APIs
type HTTPProvider struct {
	*BaseProvider
}

// NewHTTPProvider creates a new HTTP provider with the given configuration
func NewHTTPProvider(config *ProviderConfig) *HTTPProvider {
	if config == nil {
		config = DefaultConfig()
	}
	if config.Name == "" {
		config.Name = "http"
	}

	return &HTTPProvider{
		BaseProvider: NewBaseProvider(config),
	}
}

// NewHTTPProviderWithBaseURL creates a new HTTP provider with a base URL and API key
// This is a convenience method that defaults to Chat Completions API
func NewHTTPProviderWithBaseURL(baseURL, apiKey string) *HTTPProvider {
	config := NewProviderConfig("http").
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithAPIType(APITypeChatCompletions)

	return NewHTTPProvider(config)
}

// NewHTTPProviderWithBaseURLAndPath creates a new HTTP provider with base URL, API key, and base path
// The basePath is stripped from the endpoint before appending to base URL.
// For example, with baseURL="https://api.siliconflow.cn/v1" and basePath="/v1",
// the endpoint "/v1/chat/completions" becomes "https://api.siliconflow.cn/v1/chat/completions"
func NewHTTPProviderWithBaseURLAndPath(baseURL, apiKey, basePath string) *HTTPProvider {
	config := NewProviderConfig("http").
		WithBaseURL(baseURL).
		WithBasePath(basePath).
		WithAPIKey(apiKey).
		WithAPIType(APITypeChatCompletions)

	return NewHTTPProvider(config)
}

// NewHTTPProviderFull creates a new HTTP provider with full configuration options
func NewHTTPProviderFull(name, baseURL, apiKey string, supportedAPIs APIType) *HTTPProvider {
	config := NewProviderConfig(name).
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithAPIType(supportedAPIs)

	return NewHTTPProvider(config)
}

// SendRequest implements Provider.SendRequest
func (p *HTTPProvider) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	// Check if the provider supports the requested API type
	if !p.SupportedAPIs().Supports(req.APIType) {
		// Try to convert to a supported API type
		if err := p.ConvertRequestIfNeeded(req); err != nil {
			return nil, fmt.Errorf("API type %v not supported by provider %s: %w", req.APIType, p.Name(), err)
		}
	}

	// Send request using the base provider's OpenAI-compatible implementation
	return p.SendRequestToOpenAIProvider(ctx, req)
}

// SupportedAPIs returns the APIs this provider supports
func (p *HTTPProvider) SupportedAPIs() APIType {
	return p.config.SupportedAPIs
}

// Name returns the provider name
func (p *HTTPProvider) Name() string {
	return p.config.Name
}

// Ensure HTTPProvider implements Provider
var _ Provider = (*HTTPProvider)(nil)
