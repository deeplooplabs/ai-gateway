package openai

import (
	"encoding/base64"
	"encoding/json"
	"math"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Delta        *Delta  `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason"`
}

// Delta represents streaming message delta
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	N                *int      `json:"n,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Stop             any       `json:"stop,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	Tools            []Tool    `json:"tools,omitempty"`
	ToolChoice       any       `json:"tool_choice,omitempty"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ChatCompletionStreamResponse represents a streaming chunk
type ChatCompletionStreamResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Input          any    `json:"input"` // string, []string, or [][]string
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"` // "float" or "base64"
	Dimensions     int    `json:"dimensions,omitempty"`      // embedding dimensions
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding vector
type Embedding struct {
	Object    string       `json:"object"`
	Embedding EmbeddingVec `json:"embedding"`
	Index     int          `json:"index"`
}

// EmbeddingVec is an embedding vector that can be represented as either
// []float32 (float format) or a base64-encoded string (base64 format)
type EmbeddingVec []float32

// UnmarshalJSON implements json.Unmarshaler for EmbeddingVec
// It handles both []float32 arrays and base64-encoded strings
func (e *EmbeddingVec) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal as a float array
	var floats []float32
	if err := json.Unmarshal(data, &floats); err == nil {
		*e = floats
		return nil
	}

	// If that fails, try as a base64 string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Decode base64 string
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	// Convert bytes to float32 array (little-endian)
	floats = make([]float32, len(decoded)/4)
	for i := 0; i < len(floats); i++ {
		// Read 4 bytes as a little-endian float32
		bits := uint32(decoded[i*4]) | uint32(decoded[i*4+1])<<8 | uint32(decoded[i*4+2])<<16 | uint32(decoded[i*4+3])<<24
		floats[i] = math.Float32frombits(bits)
	}
	*e = floats
	return nil
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	Model   string `json:"model,omitempty"`
	Prompt  string `json:"prompt"`
	N       int    `json:"n,omitempty"`
	Size    string `json:"size,omitempty"`    // "256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"
	Quality string `json:"quality,omitempty"` // "standard" or "hd"
	Style   string `json:"style,omitempty"`   // "vivid" or "natural"
}

// ImageResponse represents an image generation response
type ImageResponse struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

// Image represents a generated image
type Image struct {
	URL           string `json:"url,omitempty"`      // For DALL-E 2
	B64JSON       string `json:"b64_json,omitempty"` // For DALL-E 3
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Type     string             `json:"type"`     // "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition defines a function tool
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ToolCall represents a tool call in a response
type ToolCall struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ToolCallChoice controls tool calling behavior
type ToolCallChoice any // Can be "none", "auto", or a specific object

// Model represents an AI model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse represents the response from listing models
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}
