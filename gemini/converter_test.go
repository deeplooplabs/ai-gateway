package gemini

import (
	"testing"

	"github.com/deeplooplabs/ai-gateway/openai"
)

func TestOpenAIToGemini(t *testing.T) {
	temp := 0.7
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: &temp,
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if len(geminiReq.Contents) != 1 {
		t.Errorf("expected 1 content, got %d", len(geminiReq.Contents))
	}
	if geminiReq.Contents[0].Role != "user" {
		t.Errorf("expected role user, got %s", geminiReq.Contents[0].Role)
	}
	if geminiReq.GenerationConfig.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", geminiReq.GenerationConfig.Temperature)
	}
}

func TestOpenAIToGeminiAssistantRole(t *testing.T) {
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
		},
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if len(geminiReq.Contents) != 2 {
		t.Errorf("expected 2 contents, got %d", len(geminiReq.Contents))
	}
	if geminiReq.Contents[1].Role != "model" {
		t.Errorf("expected role model for assistant, got %s", geminiReq.Contents[1].Role)
	}
}

func TestOpenAIToGeminiSystemRole(t *testing.T) {
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if len(geminiReq.Contents) != 2 {
		t.Errorf("expected 2 contents, got %d", len(geminiReq.Contents))
	}
	if geminiReq.Contents[0].Role != "user" {
		t.Errorf("expected role user for system, got %s", geminiReq.Contents[0].Role)
	}
}

func TestOpenAIToGeminiWithTopP(t *testing.T) {
	topP := 0.9
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "user", Content: "Hello"},
		},
		TopP: &topP,
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if geminiReq.GenerationConfig.TopP != 0.9 {
		t.Errorf("expected topP 0.9, got %f", geminiReq.GenerationConfig.TopP)
	}
}

func TestOpenAIToGeminiWithMaxTokens(t *testing.T) {
	maxTokens := 100
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: &maxTokens,
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if geminiReq.GenerationConfig.MaxOutputTokens != 100 {
		t.Errorf("expected MaxOutputTokens 100, got %d", geminiReq.GenerationConfig.MaxOutputTokens)
	}
}

func TestOpenAIToGeminiWithStopSequences(t *testing.T) {
	openaiReq := &openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.Message{
			{Role: "user", Content: "Hello"},
		},
		Stop: []string{"\n", "END"},
	}

	geminiReq := OpenAIToGemini(openaiReq, "gemini-pro")

	if len(geminiReq.GenerationConfig.StopSequences) != 2 {
		t.Errorf("expected 2 stop sequences, got %d", len(geminiReq.GenerationConfig.StopSequences))
	}
}
