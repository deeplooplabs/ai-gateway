package openresponses

import (
	"encoding/json"
	"testing"
)

func TestCreateRequestUnmarshal(t *testing.T) {
	input := `{
		"model": "gpt-4o",
		"input": "Hello, world!",
		"stream": false,
		"temperature": 0.7,
		"max_output_tokens": 1000
	}`

	var req CreateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", req.Model)
	}

	if req.Input != "Hello, world!" {
		t.Errorf("Expected input 'Hello, world!', got '%v'", req.Input)
	}

	if req.Stream == nil || *req.Stream {
		t.Error("Expected stream to be false")
	}

	if req.Temperature == nil || *req.Temperature != 0.7 {
		t.Error("Expected temperature to be 0.7")
	}

	if req.MaxOutputTokens == nil || *req.MaxOutputTokens != 1000 {
		t.Error("Expected max_output_tokens to be 1000")
	}
}

func TestCreateRequestWithArrayInput(t *testing.T) {
	input := `{
		"model": "gpt-4o",
		"input": [
			{
				"type": "message",
				"role": "user",
				"content": "Hello"
			},
			{
				"type": "message",
				"role": "assistant",
				"content": "Hi there!"
			}
		]
	}`

	var req CreateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", req.Model)
	}

	// Input should be a JSON array (comes in as []interface{})
	inputBytes, err := json.Marshal(req.Input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(inputBytes, &items); err != nil {
		t.Fatalf("Failed to parse input as array: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestResponseMarshal(t *testing.T) {
	// Create empty metadata object
	emptyMetadata := make(MetadataParam)

	// Default text format - format is required
	textFormat := &TextResponseFormat{Type: "text"}

	resp := &Response{
		ID:      "resp_abc123",
		Object:  "response",
		Status:  ResponseStatusCompleted,
		CreatedAt: 1234567890,
		CompletedAt: int64Ptr(1234567892),
		Model:   "gpt-4o",
		PreviousResponseID: nil,
		Instructions:      nil,
		Output: []ItemField{
			&MessageItem{
				ID:     "msg_xyz",
				Type:   "message",
				Status: MessageStatusCompleted,
				Role:   MessageRoleAssistant,
				Content: []OutputTextContent{
					{
						Type:        "output_text",
						Text:        "Hello, world!",
						Annotations: []Annotation{}, // Required, empty array
						Logprobs:    []LogProb{},    // Required, empty array
					},
				},
			},
		},
		Error:             nil,
		Tools:             []Tool{},
		Truncation:        TruncationAuto,
		ParallelToolCalls: true,
		Text:              TextField{Format: textFormat}, // format is required
		TopP:              1.0,
		PresencePenalty:   0.0,
		FrequencyPenalty:  0.0,
		TopLogprobs:       0,
		Temperature:       1.0,
		Reasoning:         nil,
		User:              nil,
		ToolChoice:        "auto",
		Store:             true,
		Background:        false,
		ServiceTier:       "auto",
		Metadata:          &emptyMetadata,
		Usage: &Usage{
			InputTokens:         10,
			OutputTokens:        20,
			TotalTokens:         30,
			InputTokensDetails:  &InputTokensDetails{CachedTokens: 0},
			OutputTokensDetails: &OutputTokensDetails{ReasoningTokens: 0},
		},
		MaxOutputTokens:   nil,
		MaxToolCalls:      nil,
		IncompleteDetails: nil,
		SafetyIdentifier:  nil,
		PromptCacheKey:    nil,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled Response
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != resp.ID {
		t.Errorf("Expected ID '%s', got '%s'", resp.ID, unmarshaled.ID)
	}

	if unmarshaled.Status != ResponseStatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", unmarshaled.Status)
	}

	if len(unmarshaled.Output) != 1 {
		t.Fatalf("Expected 1 output item, got %d", len(unmarshaled.Output))
	}

	// Verify required fields are present in JSON output
	jsonStr := string(data)
	requiredFields := []string{
		`"id":"resp_abc123"`,
		`"object":"response"`,
		`"status":"completed"`,
		`"created_at":1234567890`,
		`"completed_at":1234567892`,
		`"model":"gpt-4o"`,
		`"previous_response_id":null`,
		`"instructions":null`,
		`"output":[`,
		`"error":null`,
		`"tools":[]`,
		`"tool_choice":"auto"`,
		`"truncation":"auto"`,
		`"parallel_tool_calls":true`,
		`"text":{"format":{"type":"text"}}`,
		`"top_p":1`,
		`"presence_penalty":0`,
		`"frequency_penalty":0`,
		`"top_logprobs":0`,
		`"temperature":1`,
		`"reasoning":null`,
		`"user":null`,
		`"usage":{`,
		`"max_output_tokens":null`,
		`"max_tool_calls":null`,
		`"store":true`,
		`"background":false`,
		`"service_tier":"auto"`,
		`"metadata":{}`,
		`"incomplete_details":null`,
		`"safety_identifier":null`,
		`"prompt_cache_key":null`,
		`"annotations":[]`,
		`"logprobs":[]`,
	}

	for _, field := range requiredFields {
		if !contains(jsonStr, field) {
			t.Errorf("Required field %s not found in JSON output: %s", field, jsonStr)
		}
	}
}

func TestErrorMarshal(t *testing.T) {
	testErr := &Error{
		Type:    "invalid_request_error",
		Code:    "missing_model",
		Message: "model is required",
		Param:   "model",
	}

	data, err := json.Marshal(testErr)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"type":"invalid_request_error","code":"missing_model","message":"model is required","param":"model"}`
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}

	// Test in error response wrapper
	errorResp := map[string]*Error{"error": testErr}
	data, err = json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}
}

func TestToolTypes(t *testing.T) {
	input := `{
		"model": "gpt-4o",
		"input": "What's the weather?",
		"tools": [
			{
				"type": "function",
				"name": "get_weather",
				"description": "Get the current weather",
				"parameters": {
					"type": "object",
					"properties": {
						"location": {"type": "string"}
					}
				}
			}
		]
	}`

	var req CreateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(req.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(req.Tools))
	}

	tool, ok := req.Tools[0].(map[string]any)
	if !ok {
		t.Fatal("Expected tool to be a map")
	}

	if tool["type"] != "function" {
		t.Errorf("Expected tool type 'function', got '%v'", tool["type"])
	}

	if tool["name"] != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%v'", tool["name"])
	}
}

func TestTruncationEnum(t *testing.T) {
	tests := []struct {
		input    string
		expected TruncationEnum
	}{
		{"auto", TruncationAuto},
		{"disabled", TruncationDisabled},
	}

	for _, tt := range tests {
		var trunc TruncationEnum
		if err := json.Unmarshal([]byte(`"`+tt.input+`"`), &trunc); err != nil {
			t.Errorf("Failed to unmarshal %s: %v", tt.input, err)
			continue
		}
		if trunc != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, trunc)
		}
	}
}

func TestMessageRoleEnum(t *testing.T) {
	tests := []struct {
		input    string
		expected MessageRoleEnum
	}{
		{"user", MessageRoleUser},
		{"assistant", MessageRoleAssistant},
		{"system", MessageRoleSystem},
		{"developer", MessageRoleDeveloper},
	}

	for _, tt := range tests {
		var role MessageRoleEnum
		if err := json.Unmarshal([]byte(`"`+tt.input+`"`), &role); err != nil {
			t.Errorf("Failed to unmarshal %s: %v", tt.input, err)
			continue
		}
		if role != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, role)
		}
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse("resp_123", "gpt-4o")

	if resp.ID != "resp_123" {
		t.Errorf("Expected ID 'resp_123', got '%s'", resp.ID)
	}

	if resp.Object != "response" {
		t.Errorf("Expected object 'response', got '%s'", resp.Object)
	}

	if resp.Status != ResponseStatusInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", resp.Status)
	}

	if resp.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", resp.Model)
	}

	if resp.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}

	if resp.Output == nil {
		t.Error("Expected Output to be initialized")
	}

	// Verify required fields have default values
	if resp.Tools == nil {
		t.Error("Expected Tools to be initialized (required field)")
	}

	if resp.Truncation != TruncationAuto {
		t.Errorf("Expected Truncation 'auto', got '%s'", resp.Truncation)
	}

	if !resp.ParallelToolCalls {
		t.Error("Expected ParallelToolCalls to be true")
	}

	if resp.TopP != 1.0 {
		t.Errorf("Expected TopP to be 1.0, got %v", resp.TopP)
	}

	if resp.PresencePenalty != 0.0 {
		t.Errorf("Expected PresencePenalty to be 0.0, got %v", resp.PresencePenalty)
	}

	if resp.FrequencyPenalty != 0.0 {
		t.Errorf("Expected FrequencyPenalty to be 0.0, got %v", resp.FrequencyPenalty)
	}

	if resp.TopLogprobs != 0 {
		t.Errorf("Expected TopLogprobs to be 0, got %v", resp.TopLogprobs)
	}

	if !resp.Store {
		t.Error("Expected Store to be true")
	}

	if resp.Background {
		t.Error("Expected Background to be false")
	}

	if resp.ServiceTier != "auto" {
		t.Errorf("Expected ServiceTier 'auto', got '%s'", resp.ServiceTier)
	}

	// Verify nullable fields are nil
	if resp.PreviousResponseID != nil {
		t.Error("Expected PreviousResponseID to be nil")
	}

	if resp.Instructions != nil {
		t.Error("Expected Instructions to be nil")
	}

	if resp.Error != nil {
		t.Error("Expected Error to be nil")
	}

	if resp.Reasoning != nil {
		t.Error("Expected Reasoning to be nil")
	}

	if resp.User != nil {
		t.Error("Expected User to be nil")
	}

	if resp.Usage != nil {
		t.Error("Expected Usage to be nil for in_progress")
	}

	if resp.MaxOutputTokens != nil {
		t.Error("Expected MaxOutputTokens to be nil")
	}

	if resp.MaxToolCalls != nil {
		t.Error("Expected MaxToolCalls to be nil")
	}

	if resp.IncompleteDetails != nil {
		t.Error("Expected IncompleteDetails to be nil")
	}

	if resp.SafetyIdentifier != nil {
		t.Error("Expected SafetyIdentifier to be nil")
	}

	if resp.PromptCacheKey != nil {
		t.Error("Expected PromptCacheKey to be nil")
	}
}

func TestNewError(t *testing.T) {
	err := NewError("invalid_request_error", "missing_param", "param is required", "param")

	if err.Type != "invalid_request_error" {
		t.Errorf("Expected type 'invalid_request_error', got '%s'", err.Type)
	}

	if err.Code != "missing_param" {
		t.Errorf("Expected code 'missing_param', got '%s'", err.Code)
	}

	if err.Message != "param is required" {
		t.Errorf("Expected message 'param is required', got '%s'", err.Message)
	}

	if err.Param != "param" {
		t.Errorf("Expected param 'param', got '%s'", err.Param)
	}
}
