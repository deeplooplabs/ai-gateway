package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gemini2 "github.com/deeplooplabs/ai-gateway/provider/gemini"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
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

// SendRequest sends a non-streaming request via HTTP
func (p *GeminiHTTPProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	// Handle unsupported endpoints
	if endpoint == "/v1/images/generations" {
		return nil, fmt.Errorf("image generation not supported for Gemini provider")
	}

	// Handle embeddings endpoint
	if endpoint == "/v1/embeddings" {
		return p.sendEmbeddingsRequest(ctx, req)
	}

	// Handle chat completions
	return p.sendChatRequest(ctx, req)
}

func (p *GeminiHTTPProvider) sendChatRequest(ctx context.Context, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Gemini format
	model := req.Model
	if model == "" {
		model = "gemini-pro"
	}
	geminiReq := gemini2.OpenAIToGemini(req, model)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, model, p.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp gemini2.GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Convert back to OpenAI format
	return gemini2.GeminiToOpenAI(&geminiResp, model), nil
}

func (p *GeminiHTTPProvider) sendEmbeddingsRequest(ctx context.Context, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	// Parse as embedding request from the original request
	// This is a workaround - ideally we'd have separate request types
	// For now, return an error indicating this needs proper request handling
	return nil, fmt.Errorf("embeddings endpoint requires separate EmbeddingRequest type")
}
