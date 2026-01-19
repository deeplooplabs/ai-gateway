package openai

import (
	"encoding/json"
	"testing"
)

func TestChatCompletionRequest_UnmarshalJSON(t *testing.T) {
	body := `{
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Hello"}],
        "temperature": 0.7,
        "stream": false
    }`

	var req ChatCompletionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Model != "gpt-4" {
		t.Errorf("expected 'gpt-4', got '%s'", req.Model)
	}
	if len(req.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(req.Messages))
	}
	if req.Messages[0].Role != "user" {
		t.Errorf("expected 'user', got '%s'", req.Messages[0].Role)
	}
	if req.Temperature == nil || *req.Temperature != 0.7 {
		t.Error("temperature should be 0.7")
	}
}

func TestChatCompletionResponse_MarshalJSON(t *testing.T) {
	resp := &ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []Choice{{
			Index: 0,
			Message: Message{
				Role:    "assistant",
				Content: "Hello!",
			},
			FinishReason: "stop",
		}},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["id"] != "chatcmpl-123" {
		t.Errorf("expected 'chatcmpl-123', got '%v'", decoded["id"])
	}
	if decoded["object"] != "chat.completion" {
		t.Errorf("expected 'chat.completion', got '%v'", decoded["object"])
	}
}

func TestChatCompletionStreamResponse_MarshalJSON(t *testing.T) {
	resp := &ChatCompletionStreamResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []Choice{{
			Index: 0,
			Delta: &Delta{
				Content: "Hello!",
			},
			FinishReason: "",
		}},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["object"] != "chat.completion.chunk" {
		t.Errorf("expected 'chat.completion.chunk', got '%v'", decoded["object"])
	}
}

func TestUsage_MarshalJSON(t *testing.T) {
	usage := &Usage{
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
	}

	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["prompt_tokens"] != float64(10) {
		t.Errorf("expected 10, got '%v'", decoded["prompt_tokens"])
	}
	if decoded["completion_tokens"] != float64(5) {
		t.Errorf("expected 5, got '%v'", decoded["completion_tokens"])
	}
	if decoded["total_tokens"] != float64(15) {
		t.Errorf("expected 15, got '%v'", decoded["total_tokens"])
	}
}

func TestDelta_MarshalJSON(t *testing.T) {
	delta := &Delta{
		Role:    "assistant",
		Content: "Hello!",
	}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["role"] != "assistant" {
		t.Errorf("expected 'assistant', got '%v'", decoded["role"])
	}
	if decoded["content"] != "Hello!" {
		t.Errorf("expected 'Hello!', got '%v'", decoded["content"])
	}
}

func TestDelta_EmptyFieldsOmitted(t *testing.T) {
	delta := &Delta{}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if _, exists := decoded["role"]; exists {
		t.Error("expected 'role' to be omitted when empty")
	}
	if _, exists := decoded["content"]; exists {
		t.Error("expected 'content' to be omitted when empty")
	}
}

func TestEmbeddingRequest_UnmarshalJSON(t *testing.T) {
	body := `{
		"input": "hello world",
		"model": "text-embedding-3-small",
		"encoding_format": "float"
	}`

	var req EmbeddingRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Model != "text-embedding-3-small" {
		t.Errorf("expected 'text-embedding-3-small', got '%s'", req.Model)
	}
	if req.Input.(string) != "hello world" {
		t.Error("input should be 'hello world'")
	}
}

func TestImageRequest_UnmarshalJSON(t *testing.T) {
	body := `{
		"model": "dall-e-3",
		"prompt": "a cat",
		"n": 2,
		"size": "1024x1024"
	}`

	var req ImageRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Model != "dall-e-3" {
		t.Errorf("expected 'dall-e-3', got '%s'", req.Model)
	}
	if req.N != 2 {
		t.Errorf("expected n=2, got %d", req.N)
	}
}

func TestImageResponse_MarshalJSON(t *testing.T) {
	resp := &ImageResponse{
		Created: 1234567890,
		Data: []Image{{
			URL: "https://example.com/image.png",
		}},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["created"].(float64) != 1234567890 {
		t.Error("created timestamp mismatch")
	}
}
