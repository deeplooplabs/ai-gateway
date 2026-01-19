package openai

import (
	"testing"
)

func TestSSEParser_ParseLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		event  string
		data   string
		isDone bool
	}{
		{
			name:   "data line",
			line:   "data: {\"content\": \"hello\"}",
			event:  "",
			data:   "{\"content\": \"hello\"}",
			isDone: false,
		},
		{
			name:   "done marker",
			line:   "data: [DONE]",
			event:  "",
			data:   "",
			isDone: true,
		},
		{
			name:   "event line",
			line:   "event: message",
			event:  "message",
			data:   "",
			isDone: false,
		},
		{
			name:   "empty line",
			line:   "",
			event:  "",
			data:   "",
			isDone: false,
		},
		{
			name:   "comment line",
			line:   ": this is a comment",
			event:  "",
			data:   "",
			isDone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, data, isDone := ParseSSELine([]byte(tt.line))
			if event != tt.event {
				t.Errorf("expected event '%s', got '%s'", tt.event, event)
			}
			if data != tt.data {
				t.Errorf("expected data '%s', got '%s'", tt.data, data)
			}
			if isDone != tt.isDone {
				t.Errorf("expected isDone %v, got %v", tt.isDone, isDone)
			}
		})
	}
}

func TestSSEParser_IsDoneMarker(t *testing.T) {
	if !IsDoneMarker([]byte("data: [DONE]")) {
		t.Error("expected true for [DONE] marker")
	}
	if IsDoneMarker([]byte("data: something")) {
		t.Error("expected false for normal data")
	}
}

func TestSSEParser_ExtractData(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"data: {\"id\": \"123\"}", "{\"id\": \"123\"}"},
		{"data:  {\"id\": \"123\"}", "{\"id\": \"123\"}"}, // with space after colon
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractData([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
