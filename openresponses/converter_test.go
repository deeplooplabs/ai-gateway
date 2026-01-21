package openresponses

import (
	"encoding/json"
	"testing"

	openai "github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestConverter_InputToMessages_ArrayInput(t *testing.T) {
	c := NewConverter()

	// Test the exact format the user was using
	jsonInput := `[{"type":"message","role":"user","content":"Say hello in exactly 3 words."}]`

	var req CreateRequest
	if err := json.Unmarshal([]byte(`{"model":"gpt-4o","input":`+jsonInput+`}`), &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	messages, err := c.inputToMessages(req.Input)
	if err != nil {
		t.Fatalf("inputToMessages failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", messages[0].Role)
	}

	if messages[0].Content != "Say hello in exactly 3 words." {
		t.Errorf("Expected content 'Say hello in exactly 3 words.', got '%s'", messages[0].Content)
	}
}

func TestConverter_InputToMessages_StringInput(t *testing.T) {
	c := NewConverter()

	req := CreateRequest{
		Input: "Just a simple string input",
	}

	messages, err := c.inputToMessages(req.Input)
	if err != nil {
		t.Fatalf("inputToMessages failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", messages[0].Role)
	}

	if messages[0].Content != "Just a simple string input" {
		t.Errorf("Expected content 'Just a simple string input', got '%s'", messages[0].Content)
	}
}

func TestConverter_InputToMessages_MultipleMessages(t *testing.T) {
	c := NewConverter()

	jsonInput := `[
		{"type":"message","role":"system","content":"You are a helpful assistant."},
		{"type":"message","role":"user","content":"Hello!"},
		{"type":"message","role":"assistant","content":"Hi there!"},
		{"type":"message","role":"user","content":"How are you?"}
	]`

	var req CreateRequest
	if err := json.Unmarshal([]byte(`{"model":"gpt-4o","input":`+jsonInput+`}`), &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	messages, err := c.inputToMessages(req.Input)
	if err != nil {
		t.Fatalf("inputToMessages failed: %v", err)
	}

	if len(messages) != 4 {
		t.Fatalf("Expected 4 messages, got %d", len(messages))
	}

	expected := []openai.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello!"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	for i, msg := range messages {
		if msg.Role != expected[i].Role {
			t.Errorf("Message %d: expected role '%s', got '%s'", i, expected[i].Role, msg.Role)
		}
		if msg.Content != expected[i].Content {
			t.Errorf("Message %d: expected content '%s', got '%s'", i, expected[i].Content, msg.Content)
		}
	}
}

func TestConverter_RequestToChatConversion(t *testing.T) {
	c := NewConverter()

	jsonInput := `[{"type":"message","role":"user","content":"Say hello in exactly 3 words."}]`

	var req CreateRequest
	if err := json.Unmarshal([]byte(`{
		"model": "gpt-4o",
		"input": `+jsonInput+`,
		"temperature": 0.7,
		"max_output_tokens": 100
	}`), &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	chatReq, err := c.RequestToChatCompletion(&req)
	if err != nil {
		t.Fatalf("RequestToChatCompletion failed: %v", err)
	}

	if chatReq.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", chatReq.Model)
	}

	if chatReq.Temperature == nil || *chatReq.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", chatReq.Temperature)
	}

	if len(chatReq.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(chatReq.Messages))
	}

	if chatReq.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", chatReq.Messages[0].Role)
	}

	if chatReq.Messages[0].Content != "Say hello in exactly 3 words." {
		t.Errorf("Expected content 'Say hello in exactly 3 words.', got '%s'", chatReq.Messages[0].Content)
	}
}
