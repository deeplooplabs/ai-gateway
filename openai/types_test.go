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
