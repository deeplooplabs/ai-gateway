package provider

import (
	"context"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
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

// LegacyProvider is the old interface for backward compatibility
// Deprecated: Use Provider instead
type LegacyProvider interface {
	// Name returns the provider name
	Name() string
	// SendRequest sends a non-streaming request and returns the response
	SendRequest(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (*openai2.ChatCompletionResponse, error)
	// SendRequestStream sends a streaming request and returns channels for chunks and errors
	SendRequestStream(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (<-chan openai2.StreamChunk, <-chan error)
}

// LegacyProviderAdapter wraps a LegacyProvider to implement the new Provider interface
type LegacyProviderAdapter struct {
	provider LegacyProvider
	apiType  APIType
}

// NewLegacyProviderAdapter creates a new adapter for legacy providers
func NewLegacyProviderAdapter(provider LegacyProvider, apiType APIType) Provider {
	return &LegacyProviderAdapter{
		provider: provider,
		apiType:  apiType,
	}
}

// Name returns the provider name
func (a *LegacyProviderAdapter) Name() string {
	return a.provider.Name()
}

// SupportedAPIs returns the supported API types
func (a *LegacyProviderAdapter) SupportedAPIs() APIType {
	return a.apiType
}

// SendRequest implements the new Provider interface
func (a *LegacyProviderAdapter) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	// Convert to Chat Completions request
	chatReq, err := req.ToChatCompletionRequest()
	if err != nil {
		return nil, err
	}

	if req.Stream {
		// Use streaming
		chunkChan, _ := a.provider.SendRequestStream(ctx, req.Endpoint, chatReq)

		// Convert to new format
		newChunkChan := make(chan *Chunk, 16)
		newErrChan := make(chan error, 1)

		go func() {
			defer close(newChunkChan)
			defer close(newErrChan)

			for chunk := range chunkChan {
				newChunkChan <- &Chunk{
					Type:  ChunkTypeOpenAI,
					OpenAI: &chunk,
					Done:  chunk.Done,
				}
			}
		}()

		closeFn := func() error {
			// Channels will be closed by goroutine
			return nil
		}

		return NewStreamingResponse(a.apiType, newChunkChan, newErrChan, closeFn), nil
	}

	// Non-streaming
	chatResp, err := a.provider.SendRequest(ctx, req.Endpoint, chatReq)
	if err != nil {
		return nil, err
	}

	return NewChatCompletionResponse(chatResp), nil
}

// Ensure BaseProvider implements Provider
var _ Provider = (*BaseProvider)(nil)
