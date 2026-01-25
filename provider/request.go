package provider

import (
	"encoding/json"
	"fmt"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
	openresponses "github.com/deeplooplabs/ai-gateway/openresponses"
)

// Request is a unified request structure that supports Chat Completions, Responses, Embeddings, and Images APIs
type Request struct {
	// APIType specifies which API format to use (ChatCompletions, Responses, Embeddings, or Images)
	APIType APIType

	// Stream indicates whether to use streaming
	Stream bool

	// Model is the model identifier
	Model string

	// Endpoint is the specific API endpoint (e.g., "/v1/chat/completions", "/v1/responses")
	Endpoint string

	// === Chat Completions fields ===

	// Messages is for Chat Completions format
	Messages []openai.Message

	// === Responses fields ===

	// Input is for Responses format (can be string, array, or structured content)
	Input any

	// PreviousResponseID is for continuing a previous response (Responses API)
	PreviousResponseID string

	// Truncation specifies how to handle context truncation (Responses API)
	Truncation openresponses.TruncationEnum

	// === Embeddings fields ===

	// EmbeddingInput is the input text(s) to embed (string, []string, or [][]string)
	EmbeddingInput any

	// EncodingFormat is the embedding encoding format ("float" or "base64")
	EncodingFormat string

	// Dimensions is the embedding dimensions
	Dimensions int

	// === Images fields ===

	// ImagePrompt is the prompt for image generation
	ImagePrompt string

	// ImageN is the number of images to generate
	ImageN int

	// ImageSize is the image size (e.g., "1024x1024")
	ImageSize string

	// ImageQuality is the image quality ("standard" or "hd")
	ImageQuality string

	// ImageStyle is the image style ("vivid" or "natural")
	ImageStyle string

	// === Common parameters (shared by both APIs) ===

	// Temperature controls randomness
	Temperature *float64

	// TopP controls nucleus sampling
	TopP *float64

	// MaxTokens is the maximum tokens to generate
	MaxTokens *int

	// MaxOutputTokens is the maximum output tokens (Responses API naming)
	MaxOutputTokens *int

	// Stop sequences
	Stop any

	// Presence penalty
	PresencePenalty *float64

	// Frequency penalty
	FrequencyPenalty *float64

	// Tools for function calling
	Tools []openai.Tool

	// ToolChoice controls tool calling behavior
	ToolChoice any

	// Original request body for passthrough
	OriginalBody []byte

	// Additional headers for the request
	Headers map[string]string
}

// NewChatCompletionsRequest creates a new request for Chat Completions API
func NewChatCompletionsRequest(model string, messages []openai.Message) *Request {
	return &Request{
		APIType:  APITypeChatCompletions,
		Model:    model,
		Messages: messages,
		Endpoint: "/v1/chat/completions",
	}
}

// NewResponsesRequest creates a new request for Responses API
func NewResponsesRequest(model string, input any) *Request {
	return &Request{
		APIType:  APITypeResponses,
		Model:    model,
		Input:    input,
		Endpoint: "/v1/responses",
	}
}

// NewEmbeddingsRequest creates a new request for Embeddings API
func NewEmbeddingsRequest(model string, input any) *Request {
	return &Request{
		APIType:        APITypeEmbeddings,
		Model:          model,
		EmbeddingInput: input,
		Endpoint:       "/v1/embeddings",
	}
}

// NewImagesRequest creates a new request for Images API
func NewImagesRequest(model, prompt string) *Request {
	return &Request{
		APIType:     APITypeImages,
		Model:       model,
		ImagePrompt: prompt,
		Endpoint:    "/v1/images/generations",
	}
}

// GetMaxTokens returns the max tokens value, checking both field names
func (r *Request) GetMaxTokens() *int {
	if r.MaxOutputTokens != nil {
		return r.MaxOutputTokens
	}
	return r.MaxTokens
}

// SetMaxTokens sets the max tokens value
func (r *Request) SetMaxTokens(maxTokens *int) {
	r.MaxTokens = maxTokens
	r.MaxOutputTokens = maxTokens
}

// ToChatCompletionRequest converts the unified request to OpenAI ChatCompletionRequest
func (r *Request) ToChatCompletionRequest() (*openai.ChatCompletionRequest, error) {
	req := &openai.ChatCompletionRequest{
		Model:            r.Model,
		Messages:         r.Messages,
		Temperature:      r.Temperature,
		TopP:             r.TopP,
		MaxTokens:        r.GetMaxTokens(),
		Stop:             r.Stop,
		PresencePenalty:  r.PresencePenalty,
		FrequencyPenalty: r.FrequencyPenalty,
		Tools:            r.Tools,
		ToolChoice:       r.ToolChoice,
		Stream:           r.Stream,
	}
	return req, nil
}

// ToResponsesRequest converts the unified request to OpenResponses CreateRequest
func (r *Request) ToResponsesRequest() (*openresponses.CreateRequest, error) {
	req := &openresponses.CreateRequest{
		Model:              r.Model,
		Input:             r.Input,
		PreviousResponseID: r.PreviousResponseID,
		Temperature:        r.Temperature,
		TopP:              r.TopP,
		MaxOutputTokens:   r.MaxOutputTokens,
		PresencePenalty:  r.PresencePenalty,
		FrequencyPenalty: r.FrequencyPenalty,
		Truncation:       r.Truncation,
	}

	// Convert tools
	if len(r.Tools) > 0 {
		req.Tools = make([]openresponses.Tool, len(r.Tools))
		for i, tool := range r.Tools {
			req.Tools[i] = &openresponses.FunctionTool{
				Type:        tool.Type,
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
		}
	}

	if r.Stream {
		req.Stream = &r.Stream
	}

	return req, nil
}

// ToEmbeddingRequest converts the unified request to OpenAI EmbeddingRequest
func (r *Request) ToEmbeddingRequest() (*openai.EmbeddingRequest, error) {
	req := &openai.EmbeddingRequest{
		Input:          r.EmbeddingInput,
		Model:          r.Model,
		EncodingFormat: r.EncodingFormat,
		Dimensions:     r.Dimensions,
	}
	return req, nil
}

// ToImageRequest converts the unified request to OpenAI ImageRequest
func (r *Request) ToImageRequest() (*openai.ImageRequest, error) {
	req := &openai.ImageRequest{
		Model:   r.Model,
		Prompt:  r.ImagePrompt,
		N:       r.ImageN,
		Size:    r.ImageSize,
		Quality: r.ImageQuality,
		Style:   r.ImageStyle,
	}
	return req, nil
}

// Clone creates a deep copy of the request
func (r *Request) Clone() (*Request, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var cloned Request
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil, err
	}
	return &cloned, nil
}

// InputToMessages converts Responses Input to Chat Completions Messages
func (r *Request) InputToMessages() ([]openai.Message, error) {
	// If input is a string, treat it as a user message
	if str, ok := r.Input.(string); ok {
		return []openai.Message{{Role: "user", Content: str}}, nil
	}

	// If input is already a JSON array (unmarshaled as []interface{})
	if items, ok := r.Input.([]interface{}); ok {
		return messagesFromItems(items)
	}

	// If input is a JSON array (as bytes)
	if bytes, ok := r.Input.([]byte); ok {
		var items []json.RawMessage
		if err := json.Unmarshal(bytes, &items); err != nil {
			return nil, err
		}
		var itemsIfaces []interface{}
		for _, item := range items {
			itemsIfaces = append(itemsIfaces, item)
		}
		return messagesFromItems(itemsIfaces)
	}

	// Fallback: try to marshal and parse
	data, err := json.Marshal(r.Input)
	if err != nil {
		return nil, err
	}

	// Try as array first
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err == nil {
		var itemsIfaces []interface{}
		for _, item := range items {
			itemsIfaces = append(itemsIfaces, item)
		}
		return messagesFromItems(itemsIfaces)
	}

	// Treat as single user message
	return []openai.Message{{Role: "user", Content: string(data)}}, nil
}

// messagesFromItems extracts messages from input items
func messagesFromItems(items []interface{}) ([]openai.Message, error) {
	var messages []openai.Message

	for _, item := range items {
		// Try to convert item to map
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// Try JSON raw message
			if raw, ok := item.(json.RawMessage); ok {
				if err := json.Unmarshal(raw, &itemMap); err != nil {
					continue
				}
			} else {
				continue
			}
		}

		// Extract type and role
		itemType, _ := itemMap["type"].(string)
		role, _ := itemMap["role"].(string)

		if itemType == "message" && role != "" {
			// Extract content
			content := ""
			if contentVal, ok := itemMap["content"]; ok {
				content = fmt.Sprintf("%v", contentVal)
			}
			messages = append(messages, openai.Message{
				Role:    role,
				Content: content,
			})
		}
	}

	if len(messages) == 0 {
		// Fallback: no valid messages found
		return []openai.Message{{Role: "user", Content: ""}}, nil
	}

	return messages, nil
}
