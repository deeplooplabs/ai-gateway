package gemini

import (
	"github.com/deeplooplabs/ai-gateway/openai"
)

// OpenAIToGemini converts an OpenAI request to Gemini format
func OpenAIToGemini(req *openai.ChatCompletionRequest, model string) *GenerateContentRequest {
	geminiReq := &GenerateContentRequest{
		Contents:         make([]Content, 0, len(req.Messages)),
		GenerationConfig: GenerationConfig{},
	}

	// Convert messages to contents
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user" // Gemini treats system as user
		}

		content := Content{
			Role: role,
			Parts: []Part{
				{Text: msg.Content},
			},
		}
		geminiReq.Contents = append(geminiReq.Contents, content)
	}

	// Convert generation config
	if req.Temperature != nil && *req.Temperature > 0 {
		geminiReq.GenerationConfig.Temperature = *req.Temperature
	}
	if req.TopP != nil && *req.TopP > 0 {
		geminiReq.GenerationConfig.TopP = *req.TopP
	}
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		geminiReq.GenerationConfig.MaxOutputTokens = *req.MaxTokens
	}
	if req.Stop != nil {
		// Handle Stop which can be string or []string
		switch stop := req.Stop.(type) {
		case string:
			if stop != "" {
				geminiReq.GenerationConfig.StopSequences = []string{stop}
			}
		case []string:
			if len(stop) > 0 {
				geminiReq.GenerationConfig.StopSequences = stop
			}
		}
	}

	return geminiReq
}

// GeminiToOpenAI converts a Gemini response to OpenAI format
func GeminiToOpenAI(resp *GenerateContentResponse, model string) *openai.ChatCompletionResponse {
	openaiResp := &openai.ChatCompletionResponse{
		ID:      "gemini-" + model,
		Object:  "chat.completion",
		Created: 0, // Gemini doesn't provide timestamp
		Model:   model,
		Choices: make([]openai.Choice, 0, len(resp.Candidates)),
		Usage: openai.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
	}

	for _, candidate := range resp.Candidates {
		// Extract text from parts
		var content string
		for _, part := range candidate.Content.Parts {
			if content != "" && part.Text != "" {
				content += " "
			}
			content += part.Text
		}

		// Map finish reason
		finishReason := mapFinishReason(candidate.FinishReason)

		choice := openai.Choice{
			Index: candidate.Index,
			Message: openai.Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: finishReason,
		}
		openaiResp.Choices = append(openaiResp.Choices, choice)
	}

	return openaiResp
}

// mapFinishReason maps Gemini finish reasons to OpenAI format
func mapFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	default:
		return "stop"
	}
}
