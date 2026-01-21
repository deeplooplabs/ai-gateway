package provider

import (
	"encoding/json"
	"fmt"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
	openresponses "github.com/deeplooplabs/ai-gateway/openresponses"
)

// Converter handles conversion between different API formats
type Converter struct {
	supportedAPIs APIType
}

// NewConverter creates a new Converter for the given supported API types
func NewConverter(supportedAPIs APIType) *Converter {
	return &Converter{
		supportedAPIs: supportedAPIs,
	}
}

// ConvertRequest converts a request to the supported API format if needed
func (c *Converter) ConvertRequest(req *Request) error {
	// If the request API type is already supported, no conversion needed
	if c.supportedAPIs.Supports(req.APIType) {
		return nil
	}

	// Convert to the first supported API type
	switch c.supportedAPIs {
	case APITypeChatCompletions:
		return c.ResponsesToChatCompletions(req)
	case APITypeResponses:
		return c.ChatCompletionsToResponses(req)
	default:
		// Prefer Chat Completions for dual support
		return c.ResponsesToChatCompletions(req)
	}
}

// ConvertResponse converts a response to the requested API format if needed
func (c *Converter) ConvertResponse(resp *Response, requestedAPIType APIType) error {
	// If the response is already in the requested format, no conversion needed
	if resp.APIType == requestedAPIType {
		return nil
	}

	// Handle streaming responses
	if resp.Stream {
		return c.convertStreamingResponse(resp, requestedAPIType)
	}

	// Handle non-streaming responses
	return c.convertNonStreamingResponse(resp, requestedAPIType)
}

// ChatCompletionsToResponses converts a Chat Completions request to Responses format
func (c *Converter) ChatCompletionsToResponses(req *Request) error {
	// Convert messages to input format
	if len(req.Messages) > 0 {
		inputItems := make([]map[string]interface{}, 0, len(req.Messages))
		for _, msg := range req.Messages {
			item := map[string]interface{}{
				"type":    "message",
				"role":    msg.Role,
				"content": msg.Content,
			}
			inputItems = append(inputItems, item)
		}
		req.Input = inputItems
		req.Messages = nil
	}

	req.APIType = APITypeResponses
	return nil
}

// ResponsesToChatCompletions converts a Responses request to Chat Completions format
func (c *Converter) ResponsesToChatCompletions(req *Request) error {
	// Convert input to messages format
	messages, err := req.InputToMessages()
	if err != nil {
		return fmt.Errorf("convert input to messages: %w", err)
	}

	req.Messages = messages
	req.Input = nil
	req.APIType = APITypeChatCompletions
	return nil
}

// convertNonStreamingResponse converts a non-streaming response between formats
func (c *Converter) convertNonStreamingResponse(resp *Response, targetAPIType APIType) error {
	switch targetAPIType {
	case APITypeChatCompletions:
		// Convert Responses to Chat Completions
		if resp.ORResponse != nil {
			chatResp := c.responsesToChatCompletion(resp.ORResponse)
			resp.ChatCompletion = chatResp
			resp.ORResponse = nil
			resp.APIType = APITypeChatCompletions
		}
	case APITypeResponses:
		// Convert Chat Completions to Responses
		if resp.ChatCompletion != nil {
			orResp := c.chatCompletionToResponses(resp.ChatCompletion, "")
			resp.ORResponse = orResp
			resp.ChatCompletion = nil
			resp.APIType = APITypeResponses
		}
	}
	return nil
}

// convertStreamingResponse converts a streaming response between formats
func (c *Converter) convertStreamingResponse(resp *Response, targetAPIType APIType) error {
	// Create new channels for converted data
	convertedChunks := make(chan *Chunk, 16)
	convertedErrors := make(chan error, 1)

	oldClose := resp.CloseFunc
	resp.CloseFunc = func() error {
		close(convertedChunks)
		close(convertedErrors)
		if oldClose != nil {
			return oldClose()
		}
		return nil
	}

	// Start conversion goroutine
	go func() {
		defer close(convertedChunks)
		defer close(convertedErrors)

		for chunk := range resp.Chunks {
			var convertedChunk *Chunk

			switch targetAPIType {
			case APITypeChatCompletions:
				if chunk.Type == ChunkTypeOpenResponses && chunk.OREvent != nil {
					// Convert OpenResponses event to OpenAI chunk
					convertedChunk = c.orEventToOpenAIChunk(chunk.OREvent)
				} else {
					convertedChunk = chunk
				}
			case APITypeResponses:
				if chunk.Type == ChunkTypeOpenAI && chunk.OpenAI != nil {
					// Convert OpenAI chunk to OpenResponses event
					convertedChunk = c.openaiChunkToOREvent(chunk.OpenAI)
				} else {
					convertedChunk = chunk
				}
			default:
				convertedChunk = chunk
			}

			if convertedChunk != nil {
				convertedChunks <- convertedChunk
			}
		}
	}()

	resp.Chunks = convertedChunks
	resp.Errors = convertedErrors
	resp.APIType = targetAPIType

	return nil
}

// responsesToChatCompletion converts an OpenResponses response to Chat Completions format
func (c *Converter) responsesToChatCompletion(orResp *openresponses.Response) *openai.ChatCompletionResponse {
	choices := make([]openai.Choice, 0, len(orResp.Output))

	for _, item := range orResp.Output {
		if msg, ok := item.(*openresponses.MessageItem); ok {
			content := ""
			for _, c := range msg.Content {
				content += c.Text
			}

			choices = append(choices, openai.Choice{
				Index: len(choices),
				Message: openai.Message{
					Role:    string(msg.Role),
					Content: content,
				},
				FinishReason: "stop",
			})
		}
	}

	usage := openai.Usage{}
	if orResp.Usage != nil {
		usage.PromptTokens = orResp.Usage.InputTokens
		usage.CompletionTokens = orResp.Usage.OutputTokens
		usage.TotalTokens = orResp.Usage.TotalTokens
	}

	return &openai.ChatCompletionResponse{
		ID:      orResp.ID,
		Object:  "chat.completion",
		Created: orResp.CreatedAt,
		Model:   orResp.Model,
		Choices: choices,
		Usage:   usage,
	}
}

// chatCompletionToResponses converts a Chat Completions response to Responses format
func (c *Converter) chatCompletionToResponses(chatResp *openai.ChatCompletionResponse, responseID string) *openresponses.Response {
	if responseID == "" {
		responseID = "resp_" + chatResp.ID
	}

	output := make([]openresponses.ItemField, 0, len(chatResp.Choices))

	for _, choice := range chatResp.Choices {
		messageItem := &openresponses.MessageItem{
			ID:     generateMessageID(responseID, choice.Index),
			Type:   "message",
			Status: openresponses.MessageStatusCompleted,
			Role:   openresponses.MessageRoleEnum(choice.Message.Role),
			Content: []openresponses.OutputTextContent{
				{Type: "output_text", Text: choice.Message.Content},
			},
		}
		output = append(output, messageItem)
	}

	usage := &openresponses.Usage{}
	if chatResp.Usage.TotalTokens > 0 {
		usage.InputTokens = chatResp.Usage.PromptTokens
		usage.OutputTokens = chatResp.Usage.CompletionTokens
		usage.TotalTokens = chatResp.Usage.TotalTokens
	}

	return &openresponses.Response{
		ID:          responseID,
		Object:      "response",
		Status:      openresponses.ResponseStatusCompleted,
		CreatedAt:   chatResp.Created,
		CompletedAt: &chatResp.Created,
		Model:       chatResp.Model,
		Output:      output,
		Usage:       usage,
	}
}

// orEventToOpenAIChunk converts an OpenResponses streaming event to OpenAI chunk format
func (c *Converter) orEventToOpenAIChunk(event openresponses.StreamingEvent) *Chunk {
	switch e := event.(type) {
	case *openresponses.ResponseOutputTextDeltaEvent:
		// Create a ChatCompletionStreamResponse with the delta
		chatResp := openai.ChatCompletionStreamResponse{
			ID:      "chatcmpl-" + e.ItemID,
			Object:  "chat.completion.chunk",
			Model:   "",
			Created: 0,
			Choices: []openai.Choice{
				{
					Delta: &openai.Delta{
						Content: e.Delta,
					},
				},
			},
		}
		data, _ := json.Marshal(chatResp)
		return NewOpenAIChunk(data)

	case *openresponses.ResponseCompletedEvent:
		return NewOpenAIChunkDone()

	default:
		return nil
	}
}

// openaiChunkToOREvent converts an OpenAI streaming chunk to OpenResponses event format
func (c *Converter) openaiChunkToOREvent(chunk *openai.StreamChunk) *Chunk {
	if chunk.Done {
		// Create a response.completed event
		event := &openresponses.BaseStreamingEvent{
			Type:          "response.completed",
			SequenceNumber: 0,
		}
		return &Chunk{
			Type:    ChunkTypeOpenResponses,
			OREvent: event,
			Done:    true,
		}
	}

	var chatResp openai.ChatCompletionStreamResponse
	if err := json.Unmarshal(chunk.Data, &chatResp); err != nil {
		return nil
	}

	if len(chatResp.Choices) > 0 && chatResp.Choices[0].Delta != nil {
		delta := chatResp.Choices[0].Delta
		if delta.Content != "" {
			event := openresponses.NewResponseOutputTextDeltaEvent(
				0, // sequence number will be set by writer
				"", // itemID
				0,  // outputIndex
				0,  // contentIndex
				delta.Content,
			)
			return &Chunk{
				Type:    ChunkTypeOpenResponses,
				OREvent: event,
			}
		}
	}

	return nil
}

// Helper function to generate message IDs
func generateMessageID(responseID string, index int) string {
	return fmt.Sprintf("msg_%s_%d", responseID, index)
}
