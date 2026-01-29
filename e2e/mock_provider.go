package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/deeplooplabs/ai-gateway/provider"
	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// E2EMockProvider is a unified mock provider for all E2E tests
// It implements the provider.Provider interface and supports all API types
type E2EMockProvider struct {
	mu sync.Mutex

	// Configuration for responses
	chatResponse      *openai.ChatCompletionResponse
	embeddingResponse *openai.EmbeddingResponse
	imageResponse     *openai.ImageResponse

	// Streaming configuration
	streamChunks      [][]byte
	streamDelay       time.Duration

	// Error simulation
	shouldError bool
	errorMsg    string
	errorCode   int
}

// NewE2EMockProvider creates a new mock provider with default configuration
func NewE2EMockProvider() *E2EMockProvider {
	return &E2EMockProvider{
		streamDelay: 10 * time.Millisecond,
	}
}

// Name returns the provider name
func (m *E2EMockProvider) Name() string {
	return "e2e-mock"
}

// SupportedAPIs returns all API types
func (m *E2EMockProvider) SupportedAPIs() provider.APIType {
	return provider.APITypeAll
}

// SendRequest handles both streaming and non-streaming requests
func (m *E2EMockProvider) SendRequest(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for error simulation
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}

	// Handle different API types
	switch req.APIType {
	case provider.APITypeChatCompletions:
		return m.handleChatCompletion(ctx, req)
	case provider.APITypeEmbeddings:
		return m.handleEmbeddings(ctx, req)
	case provider.APITypeImages:
		return m.handleImages(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported API type: %v", req.APIType)
	}
}

// handleChatCompletion handles chat completion requests
func (m *E2EMockProvider) handleChatCompletion(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if req.Stream {
		return m.createStreamingResponse(ctx)
	}
	return m.createChatResponse(), nil
}

// handleEmbeddings handles embedding requests
func (m *E2EMockProvider) handleEmbeddings(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if m.embeddingResponse == nil {
		// Create default embedding response
		m.embeddingResponse = &openai.EmbeddingResponse{
			Object: "list",
			Data: []openai.Embedding{
				{
					Object:    "embedding",
					Embedding: make([]float32, 1536), // Default dimension
					Index:     0,
				},
			},
			Model: req.Model,
			Usage: openai.Usage{
				PromptTokens: 8,
				TotalTokens:  8,
			},
		}
	}
	return provider.NewEmbeddingResponse(m.embeddingResponse), nil
}

// handleImages handles image generation requests
func (m *E2EMockProvider) handleImages(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if m.imageResponse == nil {
		// Create default image response
		m.imageResponse = &openai.ImageResponse{
			Created: time.Now().Unix(),
			Data: []openai.Image{
				{
					URL: "https://example.com/image.png",
				},
			},
		}
	}
	return provider.NewImageResponse(m.imageResponse), nil
}

// createChatResponse creates a non-streaming chat response
func (m *E2EMockProvider) createChatResponse() *provider.Response {
	if m.chatResponse == nil {
		// Create default chat response
		m.chatResponse = &openai.ChatCompletionResponse{
			ID:      "chatcmpl-test123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []openai.Choice{
				{
					Index: 0,
					Message: openai.Message{
						Role:    "assistant",
						Content: "This is a test response from the mock provider.",
					},
					FinishReason: "stop",
				},
			},
			Usage: openai.Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}
	}
	return provider.NewChatCompletionResponse(m.chatResponse)
}

// createStreamingResponse creates a streaming chat response
func (m *E2EMockProvider) createStreamingResponse(ctx context.Context) (*provider.Response, error) {
	chunkChan := make(chan *provider.Chunk, 10)
	errorChan := make(chan error, 1)

	// Use default chunks if not configured
	chunks := m.streamChunks
	if len(chunks) == 0 {
		chunks = m.getDefaultStreamChunks()
	}

	// Launch goroutine to send chunks
	go func() {
		defer close(chunkChan)
		defer close(errorChan)

		for _, chunkData := range chunks {
			select {
			case <-ctx.Done():
				return
			case chunkChan <- provider.NewOpenAIChunk(chunkData):
				time.Sleep(m.streamDelay)
			}
		}

		// Send done marker
		chunkChan <- provider.NewOpenAIChunkDone()
	}()

	return &provider.Response{
		APIType: provider.APITypeChatCompletions,
		Stream:  true,
		Chunks:  chunkChan,
		Errors:  errorChan,
		CloseFunc: func() error {
			return nil
		},
	}, nil
}

// getDefaultStreamChunks returns default streaming chunks
func (m *E2EMockProvider) getDefaultStreamChunks() [][]byte {
	chunks := [][]byte{}

	// First chunk with role
	chunk1 := openai.ChatCompletionStreamResponse{
		ID:      "chatcmpl-stream123",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []openai.Choice{
			{
				Index: 0,
				Delta: &openai.Delta{
					Role:    "assistant",
					Content: "",
				},
			},
		},
	}
	data1, _ := json.Marshal(chunk1)
	chunks = append(chunks, data1)

	// Content chunks
	words := []string{"Hello", " from", " streaming", " response"}
	for _, word := range words {
		chunk := openai.ChatCompletionStreamResponse{
			ID:      "chatcmpl-stream123",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []openai.Choice{
				{
					Index: 0,
					Delta: &openai.Delta{
						Content: word,
					},
				},
			},
		}
		data, _ := json.Marshal(chunk)
		chunks = append(chunks, data)
	}

	// Final chunk with finish_reason
	chunkFinal := openai.ChatCompletionStreamResponse{
		ID:      "chatcmpl-stream123",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []openai.Choice{
			{
				Index:        0,
				Delta:        &openai.Delta{},
				FinishReason: "stop",
			},
		},
	}
	dataFinal, _ := json.Marshal(chunkFinal)
	chunks = append(chunks, dataFinal)

	return chunks
}

// Configuration methods (thread-safe)

// SetChatResponse sets the chat completion response
func (m *E2EMockProvider) SetChatResponse(resp *openai.ChatCompletionResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatResponse = resp
}

// SetEmbeddingResponse sets the embedding response
func (m *E2EMockProvider) SetEmbeddingResponse(resp *openai.EmbeddingResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embeddingResponse = resp
}

// SetImageResponse sets the image response
func (m *E2EMockProvider) SetImageResponse(resp *openai.ImageResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.imageResponse = resp
}

// SetStreamChunks sets custom streaming chunks
func (m *E2EMockProvider) SetStreamChunks(chunks [][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamChunks = chunks
}

// SetError enables error simulation
func (m *E2EMockProvider) SetError(shouldError bool, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
	m.errorMsg = errorMsg
}

// Reset resets all configuration to defaults
func (m *E2EMockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatResponse = nil
	m.embeddingResponse = nil
	m.imageResponse = nil
	m.streamChunks = nil
	m.shouldError = false
	m.errorMsg = ""
}
