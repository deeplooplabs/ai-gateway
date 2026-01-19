package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/deeplooplabs/ai-gateway/openai"
)

// HTTPProvider sends requests via HTTP
type HTTPProvider struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewHTTPProvider creates a new HTTP provider
func NewHTTPProvider(baseURL, apiKey string) *HTTPProvider {
	return &HTTPProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *HTTPProvider) Name() string {
	return "http"
}

// SendRequest sends a request via HTTP
func (p *HTTPProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := p.BaseURL + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp openai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &chatResp, nil
}

// SendRequestStream sends a streaming request via HTTP
func (p *HTTPProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
	chunkChan := make(chan openai.StreamChunk, 16)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		body, err := json.Marshal(req)
		if err != nil {
			errChan <- fmt.Errorf("marshal request: %w", err)
			return
		}

		url := p.BaseURL + endpoint
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errChan <- fmt.Errorf("create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := p.Client.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
			return
		}

		// Stream the response using bufio
		scanner := openai.NewSSEScanner(resp.Body)
		for scanner.Scan() {
			chunk := scanner.Chunk()
			chunkChan <- chunk
			if chunk.Done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	return chunkChan, errChan
}
