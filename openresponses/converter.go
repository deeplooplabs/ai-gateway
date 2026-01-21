package openresponses

import (
	"encoding/json"
	"fmt"

	openai "github.com/deeplooplabs/ai-gateway/provider/openai"
)

// Converter handles conversion between OpenAI and OpenResponses formats
type Converter struct{}

// NewConverter creates a new Converter
func NewConverter() *Converter {
	return &Converter{}
}

// RequestToChatCompletion converts an OpenResponses CreateRequest to an OpenAI ChatCompletionRequest
func (c *Converter) RequestToChatCompletion(req *CreateRequest) (*openai.ChatCompletionRequest, error) {
	chatReq := &openai.ChatCompletionRequest{
		Model:            req.Model,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		MaxTokens:        req.MaxOutputTokens,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
	}

	// Set stream flag
	if req.Stream != nil {
		chatReq.Stream = *req.Stream
	}

	// Convert input to OpenAI messages
	messages, err := c.inputToMessages(req.Input)
	if err != nil {
		return nil, fmt.Errorf("convert input: %w", err)
	}
	chatReq.Messages = messages

	// Convert tools
	if len(req.Tools) > 0 {
		chatReq.Tools = c.toolsToOpenAI(req.Tools)
	}

	return chatReq, nil
}

// inputToMessages converts OpenResponses input to OpenAI messages
func (c *Converter) inputToMessages(input InputParam) ([]openai.Message, error) {
	// If input is a string, treat it as a user message
	if str, ok := input.(string); ok {
		return []openai.Message{
			{Role: "user", Content: str},
		}, nil
	}

	// If input is a JSON array, parse items
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", input)), &items); err != nil {
		// Try as raw JSON bytes
		if bytes, ok := input.([]byte); ok {
			if err := json.Unmarshal(bytes, &items); err != nil {
				return nil, fmt.Errorf("parse input array: %w", err)
			}
		} else {
			return nil, fmt.Errorf("invalid input format")
		}
	}

	var messages []openai.Message
	for _, itemBytes := range items {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(itemBytes, &item); err != nil {
			continue
		}

		// Get type field
		var itemType string
		if typeBytes, ok := item["type"]; ok {
			json.Unmarshal(typeBytes, &itemType)
		}

		// Get role field (for message items)
		var role string
		if roleBytes, ok := item["role"]; ok {
			json.Unmarshal(roleBytes, &role)
		}

		// Get content field
		var content interface{}
		if contentBytes, ok := item["content"]; ok {
			if string(contentBytes) == "null" || len(contentBytes) == 0 {
				content = ""
			} else {
				// Try as string first
				var str string
				if err := json.Unmarshal(contentBytes, &str); err == nil {
					content = str
				} else {
					// Try as array
					var arr []json.RawMessage
					if err := json.Unmarshal(contentBytes, &arr); err == nil {
						// For array content, extract text from input_text items
						content = c.extractContentText(arr)
					}
				}
			}
		}

		if itemType == "message" && role != "" {
			messages = append(messages, openai.Message{
				Role:    role,
				Content: fmt.Sprintf("%v", content),
			})
		}
	}

	if len(messages) == 0 {
		// Fallback: treat entire input as user message
		return []openai.Message{{Role: "user", Content: fmt.Sprintf("%v", input)}}, nil
	}

	return messages, nil
}

// extractContentText extracts text from content items
func (c *Converter) extractContentText(contentItems []json.RawMessage) string {
	var result string
	for _, itemBytes := range contentItems {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(itemBytes, &item); err != nil {
			continue
		}

		var itemType string
		if typeBytes, ok := item["type"]; ok {
			json.Unmarshal(typeBytes, &itemType)
		}

		if itemType == "input_text" {
			var text string
			if textBytes, ok := item["text"]; ok {
				json.Unmarshal(textBytes, &text)
				result += text
			}
		}
	}
	return result
}

// toolsToOpenAI converts OpenResponses tools to OpenAI tools
func (c *Converter) toolsToOpenAI(tools []Tool) []openai.Tool {
	var openAITools []openai.Tool
	for _, tool := range tools {
		if fn, ok := tool.(*FunctionTool); ok {
			openAITools = append(openAITools, openai.Tool{
				Type: "function",
				Function: openai.FunctionDefinition{
					Name:        fn.Name,
					Description: fn.Description,
					Parameters:  fn.Parameters,
				},
			})
		}
	}
	return openAITools
}

// ChatCompletionToResponse converts an OpenAI ChatCompletionResponse to an OpenResponses Response
func (c *Converter) ChatCompletionToResponse(chatResp *openai.ChatCompletionResponse, responseID string) *Response {
	output := make([]ItemField, 0, len(chatResp.Choices))

	for _, choice := range chatResp.Choices {
		if choice.Message.Role == "" {
			continue
		}

		messageItem := &MessageItem{
			ID:     generateMessageID(responseID, choice.Index),
			Type:   "message",
			Status: MessageStatusCompleted,
			Role:   MessageRoleEnum(choice.Message.Role),
			Content: []OutputTextContent{
				{
					Type: "output_text",
					Text: choice.Message.Content,
				},
			},
		}
		output = append(output, messageItem)
	}

	resp := &Response{
		ID:          responseID,
		Object:      "response",
		Status:      ResponseStatusCompleted,
		CreatedAt:   chatResp.Created,
		CompletedAt: &chatResp.Created,
		Model:       chatResp.Model,
		Output:      output,
	}

	// Convert usage
	if chatResp.Usage.TotalTokens > 0 {
		resp.Usage = &Usage{
			InputTokens:  chatResp.Usage.PromptTokens,
			OutputTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:  chatResp.Usage.TotalTokens,
		}
	}

	return resp
}

// StreamingChunkToEvents converts an OpenAI streaming chunk to OpenResponses streaming events
func (c *Converter) StreamingChunkToEvents(chunk []byte, seq *int, itemID string, outputIndex int) []StreamingEvent {
	var chatResp openai.ChatCompletionStreamResponse
	if err := json.Unmarshal(chunk, &chatResp); err != nil {
		return nil
	}

	var events []StreamingEvent

	for _, choice := range chatResp.Choices {
		if choice.Delta != nil {
			// Text delta
			if choice.Delta.Content != "" {
				*seq++
				events = append(events, NewResponseOutputTextDeltaEvent(
					*seq, itemID, outputIndex, 0, choice.Delta.Content,
				))
			}
		}

		// Check if choice is complete
		if choice.FinishReason != "" {
			*seq++
			// Send done event for the content
			fullText := c.getAccumulatedText(choice)
			events = append(events, NewResponseOutputTextDoneEvent(
				*seq, itemID, outputIndex, 0, fullText,
			))

			*seq++
			// Send item done event
			messageItem := &MessageItem{
				ID:     itemID,
				Type:   "message",
				Status: MessageStatusCompleted,
				Role:   MessageRoleAssistant,
				Content: []OutputTextContent{
					{Type: "output_text", Text: fullText},
				},
			}
			events = append(events, NewResponseOutputItemDoneEvent(*seq, outputIndex, messageItem))
		}
	}

	return events
}

// getAccumulatedText extracts the accumulated text from a choice
func (c *Converter) getAccumulatedText(choice openai.Choice) string {
	if choice.Message.Content != "" {
		return choice.Message.Content
	}
	if choice.Delta != nil && choice.Delta.Content != "" {
		return choice.Delta.Content
	}
	return ""
}

// generateMessageID generates a unique message ID
func generateMessageID(responseID string, index int) string {
	return fmt.Sprintf("msg_%s_%d", responseID, index)
}

// ResponseToChatCompletion converts an OpenResponses Response to an OpenAI ChatCompletionResponse
func (c *Converter) ResponseToChatCompletion(orResp *Response) *openai.ChatCompletionResponse {
	if orResp == nil || len(orResp.Output) == 0 {
		return nil
	}

	choices := make([]openai.Choice, 0, len(orResp.Output))

	for i, item := range orResp.Output {
		if msgItem, ok := item.(*MessageItem); ok && msgItem.Role != "" {
			// Extract content text from OutputTextContent
			var content string
			for _, c := range msgItem.Content {
				if c.Type == "output_text" {
					content += c.Text
				}
			}

			// Map status to finish reason
			finishReason := "stop"
			if msgItem.Status == MessageStatusIncomplete {
				finishReason = "length"
			}

			choices = append(choices, openai.Choice{
				Index: i,
				Message: openai.Message{
					Role:    string(msgItem.Role),
					Content: content,
				},
				FinishReason: finishReason,
			})
		}
	}

	if len(choices) == 0 {
		return nil
	}

	chatResp := &openai.ChatCompletionResponse{
		ID:      orResp.ID,
		Object:  "chat.completion",
		Created: orResp.CreatedAt,
		Model:   orResp.Model,
		Choices: choices,
	}

	// Convert usage
	if orResp.Usage != nil {
		chatResp.Usage = openai.Usage{
			PromptTokens:     orResp.Usage.InputTokens,
			CompletionTokens: orResp.Usage.OutputTokens,
			TotalTokens:      orResp.Usage.TotalTokens,
		}
	}

	return chatResp
}
