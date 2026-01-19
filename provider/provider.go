package provider

import (
	"context"

	"github.com/deeplooplabs/ai-gateway/openai"
)

// Provider defines the interface for sending requests to LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string
	// SendRequest sends a non-streaming request and returns the response
	SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	// SendRequestStream sends a streaming request and returns channels for chunks and errors
	SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
}
