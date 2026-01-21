package provider

import (
	"fmt"
	"net/http"
	"time"
)

// APIType represents the API format a provider supports
type APIType int

const (
	// APITypeChatCompletions is OpenAI Chat Completions API
	APITypeChatCompletions APIType = 1 << iota
	// APITypeResponses is OpenResponses API
	APITypeResponses
	// APITypeAll supports both APIs
	APITypeAll = APITypeChatCompletions | APITypeResponses
)

// String returns the string representation of APIType
func (a APIType) String() string {
	switch a {
	case APITypeChatCompletions:
		return "chat_completions"
	case APITypeResponses:
		return "responses"
	case APITypeAll:
		return "all"
	default:
		return fmt.Sprintf("unknown(%d)", a)
	}
}

// Supports checks if the provider supports the given API type
func (a APIType) Supports(apiType APIType) bool {
	return a&apiType != 0
}

// ProviderConfig contains provider configuration
type ProviderConfig struct {
	// Name is the provider name
	Name string

	// BaseURL is the base URL for the provider API
	BaseURL string

	// APIKey is the authentication key
	APIKey string

	// SupportedAPIs is the API type(s) this provider supports
	SupportedAPIs APIType

	// HTTPClient is the HTTP client to use (optional)
	HTTPClient *http.Client

	// Timeout is the request timeout (optional)
	Timeout time.Duration

	// RequestConverter is an optional custom request converter
	RequestConverter RequestConverterFunc

	// ResponseConverter is an optional custom response converter
	ResponseConverter ResponseConverterFunc
}

// RequestConverterFunc is a function that converts a request to a supported format
type RequestConverterFunc func(*Request) error

// ResponseConverterFunc is a function that converts a response from a supported format
type ResponseConverterFunc func(*Response) error

// DefaultConfig returns a default provider configuration
func DefaultConfig() *ProviderConfig {
	return &ProviderConfig{
		SupportedAPIs: APITypeChatCompletions,
		Timeout:       60 * time.Second,
	}
}

// NewProviderConfig creates a new provider configuration with the given name
func NewProviderConfig(name string) *ProviderConfig {
	return &ProviderConfig{
		Name:          name,
		SupportedAPIs: APITypeChatCompletions,
		Timeout:       60 * time.Second,
	}
}

// WithAPIType sets the supported API types
func (c *ProviderConfig) WithAPIType(apiType APIType) *ProviderConfig {
	c.SupportedAPIs = apiType
	return c
}

// WithBaseURL sets the base URL
func (c *ProviderConfig) WithBaseURL(baseURL string) *ProviderConfig {
	c.BaseURL = baseURL
	return c
}

// WithAPIKey sets the API key
func (c *ProviderConfig) WithAPIKey(apiKey string) *ProviderConfig {
	c.APIKey = apiKey
	return c
}

// WithTimeout sets the timeout
func (c *ProviderConfig) WithTimeout(timeout time.Duration) *ProviderConfig {
	c.Timeout = timeout
	return c
}

// WithHTTPClient sets the HTTP client
func (c *ProviderConfig) WithHTTPClient(client *http.Client) *ProviderConfig {
	c.HTTPClient = client
	return c
}

// WithRequestConverter sets the request converter
func (c *ProviderConfig) WithRequestConverter(fn RequestConverterFunc) *ProviderConfig {
	c.RequestConverter = fn
	return c
}

// WithResponseConverter sets the response converter
func (c *ProviderConfig) WithResponseConverter(fn ResponseConverterFunc) *ProviderConfig {
	c.ResponseConverter = fn
	return c
}

// GetHTTPClient returns the HTTP client, creating a default one if not set
func (c *ProviderConfig) GetHTTPClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{
		Timeout: c.Timeout,
	}
}
