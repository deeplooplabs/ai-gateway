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
