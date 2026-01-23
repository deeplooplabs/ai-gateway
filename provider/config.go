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

	// BasePath is the path prefix to strip from the endpoint before appending to BaseURL.
	// For example, if BaseURL is "https://api.siliconflow.cn/v1" and BasePath is "/v1",
	// then endpoint "/v1/chat/completions" will become "https://api.siliconflow.cn/v1/chat/completions"
	// (the "/v1" prefix from the endpoint is stripped).
	// Default is empty (no stripping).
	BasePath string

	// APIKey is the authentication key
	APIKey string

	// SupportedAPIs is the API type(s) this provider supports
	SupportedAPIs APIType

	// HTTPClient is the HTTP client to use (optional)
	HTTPClient *http.Client

	// Timeout is the total request timeout (optional, default: 60s)
	Timeout time.Duration

	// ConnectTimeout is the connection timeout (optional, default: 10s)
	ConnectTimeout time.Duration

	// ReadTimeout is the read timeout (optional, default: 30s)
	ReadTimeout time.Duration

	// ConnectionPool settings
	MaxIdleConns        int           // Maximum idle connections (default: 100)
	MaxConnsPerHost     int           // Maximum connections per host (default: 10)
	IdleConnTimeout     time.Duration // Idle connection timeout (default: 90s)
	MaxIdleConnsPerHost int           // Maximum idle connections per host (default: 10)

	// Retry configuration
	RetryConfig *RetryConfig

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
		SupportedAPIs:       APITypeChatCompletions,
		Timeout:             60 * time.Second,
		ConnectTimeout:      10 * time.Second,
		ReadTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 10,
		RetryConfig:         DefaultRetryConfig(),
	}
}

// NewProviderConfig creates a new provider configuration with the given name
func NewProviderConfig(name string) *ProviderConfig {
	return &ProviderConfig{
		Name:                name,
		SupportedAPIs:       APITypeChatCompletions,
		Timeout:             60 * time.Second,
		ConnectTimeout:      10 * time.Second,
		ReadTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 10,
		RetryConfig:         DefaultRetryConfig(),
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

// WithBasePath sets the base path to strip from endpoints
func (c *ProviderConfig) WithBasePath(basePath string) *ProviderConfig {
	c.BasePath = basePath
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

// WithConnectTimeout sets the connection timeout
func (c *ProviderConfig) WithConnectTimeout(timeout time.Duration) *ProviderConfig {
	c.ConnectTimeout = timeout
	return c
}

// WithReadTimeout sets the read timeout
func (c *ProviderConfig) WithReadTimeout(timeout time.Duration) *ProviderConfig {
	c.ReadTimeout = timeout
	return c
}

// WithConnectionPool sets the connection pool parameters
func (c *ProviderConfig) WithConnectionPool(maxIdleConns, maxConnsPerHost, maxIdleConnsPerHost int, idleConnTimeout time.Duration) *ProviderConfig {
	c.MaxIdleConns = maxIdleConns
	c.MaxConnsPerHost = maxConnsPerHost
	c.MaxIdleConnsPerHost = maxIdleConnsPerHost
	c.IdleConnTimeout = idleConnTimeout
	return c
}

// WithRetryConfig sets the retry configuration
func (c *ProviderConfig) WithRetryConfig(retryConfig *RetryConfig) *ProviderConfig {
	c.RetryConfig = retryConfig
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
	
	// Create custom transport with connection pool settings
	transport := &http.Transport{
		MaxIdleConns:        c.MaxIdleConns,
		MaxConnsPerHost:     c.MaxConnsPerHost,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		IdleConnTimeout:     c.IdleConnTimeout,
		DisableKeepAlives:   false,
	}
	
	// Set timeouts if configured
	if c.ConnectTimeout > 0 {
		transport.DialContext = (&http.Transport{}).DialContext
		transport.ResponseHeaderTimeout = c.ReadTimeout
	}
	
	return &http.Client{
		Timeout:   c.Timeout,
		Transport: transport,
	}
}
