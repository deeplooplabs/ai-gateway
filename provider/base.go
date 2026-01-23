package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	config    *ProviderConfig
	client    *http.Client
	converter *Converter
}

// NewBaseProvider creates a new BaseProvider with the given configuration
func NewBaseProvider(config *ProviderConfig) *BaseProvider {
	if config == nil {
		config = DefaultConfig()
	}

	return &BaseProvider{
		config:    config,
		client:    config.GetHTTPClient(),
		converter: NewConverter(config.SupportedAPIs),
	}
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	if p.config.Name != "" {
		return p.config.Name
	}
	return "base"
}

// SupportedAPIs returns the APIs this provider supports
func (p *BaseProvider) SupportedAPIs() APIType {
	return p.config.SupportedAPIs
}

// Config returns the provider configuration
func (p *BaseProvider) Config() *ProviderConfig {
	return p.config
}

// Converter returns the converter
func (p *BaseProvider) Converter() *Converter {
	return p.converter
}

// SendRequest implements Provider.SendRequest
func (p *BaseProvider) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	return p.SendRequestToOpenAIProvider(ctx, req)
}

// sendHTTP sends an HTTP request with common headers
func (p *BaseProvider) sendHTTP(ctx context.Context, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header if API key is configured
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	// Set additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return p.client.Do(req)
}

// sendHTTPNonStreaming sends a non-streaming HTTP request
func (p *BaseProvider) sendHTTPNonStreaming(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	resp, err := p.sendHTTP(ctx, url, body, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

// ConvertRequestIfNeeded converts the request to a supported API format if needed
func (p *BaseProvider) ConvertRequestIfNeeded(req *Request) error {
	// Use custom converter if provided
	if p.config.RequestConverter != nil {
		return p.config.RequestConverter(req)
	}

	// Use default converter
	return p.converter.ConvertRequest(req)
}

// ConvertResponseIfNeeded converts the response to a requested API format if needed
func (p *BaseProvider) ConvertResponseIfNeeded(resp *Response, requestedAPIType APIType) error {
	// Use custom converter if provided
	if p.config.ResponseConverter != nil {
		return p.config.ResponseConverter(resp)
	}

	// Use default converter
	return p.converter.ConvertResponse(resp, requestedAPIType)
}

// ParseChatCompletionRequest parses the unified request as a Chat Completions request
func (p *BaseProvider) ParseChatCompletionRequest(req *Request) (*openai.ChatCompletionRequest, error) {
	return req.ToChatCompletionRequest()
}

// SendRequestToOpenAIProvider sends a request to an OpenAI-compatible provider
// This is a helper method for providers that use the OpenAI API format
func (p *BaseProvider) SendRequestToOpenAIProvider(ctx context.Context, req *Request) (*Response, error) {
	// Convert to Chat Completions format if needed
	if req.APIType != APITypeChatCompletions {
		if err := p.ConvertRequestIfNeeded(req); err != nil {
			return nil, fmt.Errorf("convert request: %w", err)
		}
	}

	// Parse as Chat Completions request
	chatReq, err := p.ParseChatCompletionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("parse chat completion request: %w", err)
	}

	// Marshal request
	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Build URL
	endpoint := req.Endpoint
	if endpoint == "" {
		endpoint = "/v1/chat/completions"
	}

	// Strip BasePath prefix from endpoint if configured
	if p.config.BasePath != "" && len(endpoint) >= len(p.config.BasePath) {
		if endpoint[:len(p.config.BasePath)] == p.config.BasePath {
			endpoint = endpoint[len(p.config.BasePath):]
			// Ensure endpoint starts with /
			if len(endpoint) > 0 && endpoint[0] != '/' {
				endpoint = "/" + endpoint
			}
		}
	}

	url := p.config.BaseURL + endpoint

	// Set headers
	headers := req.Headers
	if headers == nil {
		headers = make(map[string]string)
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		return p.sendStreamingRequest(ctx, url, body, headers, req.APIType)
	}

	return p.sendNonStreamingRequest(ctx, url, body, headers)
}

// sendNonStreamingRequest sends a non-streaming request
func (p *BaseProvider) sendNonStreamingRequest(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	respBody, err := p.sendHTTPNonStreaming(ctx, url, body, headers)
	if err != nil {
		return nil, err
	}

	var chatResp openai.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return NewChatCompletionResponse(&chatResp), nil
}

// sendStreamingRequest sends a streaming request
func (p *BaseProvider) sendStreamingRequest(ctx context.Context, url string, body []byte, headers map[string]string, apiType APIType) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}
	req.Header.Set("Accept", "text/event-stream")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	chunkChan := make(chan *Chunk, 16)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)
		defer resp.Body.Close()

		// Read SSE line by line
		decoder := NewSSEDecoder(resp.Body)
		for {
			// Check for context cancellation before reading
			if ctx.Err() != nil {
				return
			}

			line, err := decoder.NextLine()
			if err != nil {
				if err != io.EOF {
					errChan <- fmt.Errorf("read stream: %w", err)
				}
				return
			}

			// Check for [DONE]
			if openai.IsDoneMarker(line) {
				chunkChan <- NewOpenAIChunkDone()
				return
			}

			// Extract data: content
			_, data, isDone := openai.ParseSSELine(line)
			if isDone {
				chunkChan <- NewOpenAIChunkDone()
				return
			}
			if data != "" {
				chunkChan <- NewOpenAIChunk([]byte(data))
			}
		}
	}()

	closeFn := func() error {
		// The goroutine will close the body when done
		return nil
	}

	return NewStreamingResponse(apiType, chunkChan, errChan, closeFn), nil
}

// SSEDecoder helps decode SSE streams
type SSEDecoder struct {
	reader *bufio.Reader
}

// NewSSEDecoder creates a new SSE decoder
func NewSSEDecoder(r io.Reader) *SSEDecoder {
	return &SSEDecoder{
		reader: bufio.NewReader(r),
	}
}

// NextLine reads the next line from the SSE stream
func (d *SSEDecoder) NextLine() ([]byte, error) {
	return d.reader.ReadBytes('\n')
}
