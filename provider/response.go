package provider

import (
	"fmt"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
	openresponses "github.com/deeplooplabs/ai-gateway/openresponses"
)

// ChunkType represents the format of a streaming chunk
type ChunkType int

const (
	// ChunkTypeOpenAI is an OpenAI format streaming chunk
	ChunkTypeOpenAI ChunkType = iota
	// ChunkTypeOpenResponses is an OpenResponses format streaming event
	ChunkTypeOpenResponses
)

// Chunk represents a unified streaming chunk that can be either format
type Chunk struct {
	// Type indicates the chunk format
	Type ChunkType

	// OpenAI format chunk
	OpenAI *openai2.StreamChunk

	// OpenResponses format event
	OREvent openresponses.StreamingEvent

	// Done indicates the stream is complete
	Done bool
}

// NewOpenAIChunk creates a new OpenAI format chunk
func NewOpenAIChunk(data []byte) *Chunk {
	return &Chunk{
		Type:  ChunkTypeOpenAI,
		OpenAI: &openai2.StreamChunk{Data: data},
	}
}

// NewOpenAIChunkDone creates a new OpenAI format done chunk
func NewOpenAIChunkDone() *Chunk {
	return &Chunk{
		Type:  ChunkTypeOpenAI,
		OpenAI: &openai2.StreamChunk{Done: true},
		Done:  true,
	}
}

// NewOREventsChunk creates a new OpenResponses format event chunk
func NewOREventsChunk(event openresponses.StreamingEvent) *Chunk {
	return &Chunk{
		Type:    ChunkTypeOpenResponses,
		OREvent: event,
	}
}

// Response is a unified response structure that can handle both streaming and non-streaming responses
type Response struct {
	// APIType indicates the format of this response
	APIType APIType

	// Stream indicates if this is a streaming response
	Stream bool

	// === Non-streaming responses (when Stream=false) ===

	// ChatCompletion is the OpenAI Chat Completions response
	ChatCompletion *openai2.ChatCompletionResponse

	// ORResponse is the OpenResponses response
	ORResponse *openresponses.Response

	// === Streaming responses (when Stream=true) ===

	// Chunks is the channel for streaming chunks
	Chunks <-chan *Chunk

	// Errors is the channel for streaming errors
	Errors <-chan error

	// CloseFunc is called to close the streaming response
	CloseFunc func() error
}

// NewChatCompletionResponse creates a new non-streaming Chat Completions response
func NewChatCompletionResponse(resp *openai2.ChatCompletionResponse) *Response {
	return &Response{
		APIType:       APITypeChatCompletions,
		Stream:        false,
		ChatCompletion: resp,
	}
}

// NewResponsesResponse creates a new non-streaming Responses response
func NewResponsesResponse(resp *openresponses.Response) *Response {
	return &Response{
		APIType:    APITypeResponses,
		Stream:     false,
		ORResponse: resp,
	}
}

// NewStreamingResponse creates a new streaming response
func NewStreamingResponse(apiType APIType, chunks <-chan *Chunk, errors <-chan error, closeFn func() error) *Response {
	return &Response{
		APIType:   apiType,
		Stream:    true,
		Chunks:    chunks,
		Errors:    errors,
		CloseFunc: closeFn,
	}
}

// Close closes the streaming response
func (r *Response) Close() error {
	if r.CloseFunc != nil {
		return r.CloseFunc()
	}
	return nil
}

// GetChatCompletion returns the Chat Completions response, converting from Responses if needed
func (r *Response) GetChatCompletion() (*openai2.ChatCompletionResponse, error) {
	if r.ChatCompletion != nil {
		return r.ChatCompletion, nil
	}
	if r.ORResponse != nil {
		converter := openresponses.NewConverter()
		chatResp := converter.ResponseToChatCompletion(r.ORResponse)
		if chatResp == nil {
			return nil, fmt.Errorf("failed to convert ORResponse to ChatCompletion")
		}
		return chatResp, nil
	}
	return nil, fmt.Errorf("no response data available")
}

// GetORResponse returns the OpenResponses response, converting from Chat Completions if needed
func (r *Response) GetORResponse(responseID string) (*openresponses.Response, error) {
	if r.ORResponse != nil {
		return r.ORResponse, nil
	}
	if r.ChatCompletion != nil {
		converter := openresponses.NewConverter()
		return converter.ChatCompletionToResponse(r.ChatCompletion, responseID), nil
	}
	return nil, fmt.Errorf("no response data available")
}

// IsStreaming returns true if this is a streaming response
func (r *Response) IsStreaming() bool {
	return r.Stream
}

// GetAPIType returns the API type of this response
func (r *Response) GetAPIType() APIType {
	return r.APIType
}
