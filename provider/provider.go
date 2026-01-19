package provider

import (
	"context"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
)

// Provider defines the interface for sending requests to LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string
	// SendRequest sends a non-streaming request and returns the response
	SendRequest(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (*openai2.ChatCompletionResponse, error)
	// SendRequestStream sends a streaming request and returns channels for chunks and errors
	SendRequestStream(ctx context.Context, endpoint string, req *openai2.ChatCompletionRequest) (<-chan openai2.StreamChunk, <-chan error)
}
