package gemini

import (
	"testing"
)

func TestContent_Role(t *testing.T) {
	tests := []struct {
		name     string
		content  Content
		expected string
	}{
		{
			name:     "user role",
			content:  Content{Role: "user"},
			expected: "user",
		},
		{
			name:     "model role",
			content:  Content{Role: "model"},
			expected: "model",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content.Role != tt.expected {
				t.Errorf("expected role %s, got %s", tt.expected, tt.content.Role)
			}
		})
	}
}
