package provider

import (
	"net/http"
	"time"
)

// GeminiHTTPProvider sends requests to Gemini via HTTP
type GeminiHTTPProvider struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewGeminiHTTPProvider creates a new Gemini HTTP provider
func NewGeminiHTTPProvider(apiKey string) *GeminiHTTPProvider {
	return &GeminiHTTPProvider{
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *GeminiHTTPProvider) Name() string {
	return "gemini-http"
}
