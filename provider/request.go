package provider

import (
	"encoding/json"
	"fmt"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
	openresponses "github.com/deeplooplabs/ai-gateway/openresponses"
)

// Request is a unified request structure that supports both Chat Completions and Responses APIs
type Request struct {
	// APIType specifies which API format to use (ChatCompletions or Responses)
	APIType APIType

	// Stream indicates whether to use streaming
	Stream bool

	// Model is the model identifier
	Model string

	// Endpoint is the specific API endpoint (e.g., "/v1/chat/completions", "/v1/responses")
	Endpoint string

	// === Chat Completions fields ===

	// Messages is for Chat Completions format
	Messages []openai2.Message

	// === Responses fields ===

	// Input is for Responses format (can be string, array, or structured content)
	Input any

	// PreviousResponseID is for continuing a previous response (Responses API)
	PreviousResponseID string

	// Truncation specifies how to handle context truncation (Responses API)
	Truncation openresponses.TruncationEnum

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
	Tools []openai2.Tool

	// ToolChoice controls tool calling behavior
	ToolChoice any

	// Original request body for passthrough
	OriginalBody []byte

	// Additional headers for the request
	Headers map[string]string
}

// NewChatCompletionsRequest creates a new request for Chat Completions API
func NewChatCompletionsRequest(model string, messages []openai2.Message) *Request {
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
func (r *Request) ToChatCompletionRequest() (*openai2.ChatCompletionRequest, error) {
	req := &openai2.ChatCompletionRequest{
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
func (r *Request) InputToMessages() ([]openai2.Message, error) {
	// If input is a string, treat it as a user message
	if str, ok := r.Input.(string); ok {
		return []openai2.Message{{Role: "user", Content: str}}, nil
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
	return []openai2.Message{{Role: "user", Content: string(data)}}, nil
}

// messagesFromItems extracts messages from input items
func messagesFromItems(items []interface{}) ([]openai2.Message, error) {
	var messages []openai2.Message

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
			messages = append(messages, openai2.Message{
				Role:    role,
				Content: content,
			})
		}
	}

	if len(messages) == 0 {
		// Fallback: no valid messages found
		return []openai2.Message{{Role: "user", Content: ""}}, nil
	}

	return messages, nil
}
