package openresponses

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

type mockFlusher struct {
	flushed bool
}

func (m *mockFlusher) Flush() {
	m.flushed = true
}

func TestStreamWriter_WriteEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    StreamingEvent
		contains []string
	}{
		{
			name: "ResponseCreatedEvent",
			event: NewResponseCreatedEvent(1, NewResponse("resp_123", "gpt-4o")),
			contains: []string{
				"event: response.created",
				`"sequence_number":1`,
				`"id":"resp_123"`,
				`"response":{`,
			},
		},
		{
			name: "ResponseInProgressEvent",
			event: NewResponseInProgressEvent(2, NewResponse("resp_123", "gpt-4o")),
			contains: []string{
				"event: response.in_progress",
				`"sequence_number":2`,
				`"response":{`,
			},
		},
		{
			name: "ResponseOutputItemAddedEvent",
			event: NewResponseOutputItemAddedEvent(3, 0, &MessageItem{
				ID:     "msg_123",
				Type:   "message",
				Status: MessageStatusInProgress,
				Role:   MessageRoleAssistant,
				Content: []OutputTextContent{{Type: "output_text", Text: ""}},
			}),
			contains: []string{
				"event: response.output_item.added",
				`"sequence_number":3`,
				`"output_index":0`,
				`"type":"message"`,
			},
		},
		{
			name: "ResponseOutputTextDeltaEvent",
			event: NewResponseOutputTextDeltaEvent(4, "msg_123", 0, 0, "Hello"),
			contains: []string{
				"event: response.output_text.delta",
				`"sequence_number":4`,
				`"item_id":"msg_123"`,
				`"delta":"Hello"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			flusher := &mockFlusher{}
			writer := NewStreamWriter(&buf, flusher)

			err := writer.WriteEvent(tt.event)
			if err != nil {
				t.Fatalf("WriteEvent failed: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}

			if !flusher.flushed {
				t.Error("Expected flusher to be flushed")
			}
		})
	}
}

func TestStreamWriter_WriteDone(t *testing.T) {
	var buf bytes.Buffer
	flusher := &mockFlusher{}
	writer := NewStreamWriter(&buf, flusher)

	err := writer.WriteDone()
	if err != nil {
		t.Fatalf("WriteDone failed: %v", err)
	}

	output := buf.String()
	expected := "data: [DONE]\n\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	if !flusher.flushed {
		t.Error("Expected flusher to be flushed")
	}
}

func TestStreamWriter_WriteError(t *testing.T) {
	var buf bytes.Buffer
	flusher := &mockFlusher{}
	writer := NewStreamWriter(&buf, flusher)

	err := writer.WriteError(NewError("server_error", "test_error", "Test error message", ""))
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "event: error") {
		t.Error("Expected error event type")
	}
	if !contains(output, "Test error message") {
		t.Error("Expected error message in output")
	}
	if !contains(output, "data: [DONE]") {
		t.Error("Expected DONE marker after error")
	}
}

func TestStreamWriter_WithHTTPResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewStreamWriter(recorder, recorder)

	event := NewResponseCreatedEvent(1, NewResponse("resp_test", "gpt-4o"))
	err := writer.WriteEvent(event)
	if err != nil {
		t.Fatalf("WriteEvent failed: %v", err)
	}

	output := recorder.Body.String()
	if !contains(output, "response.created") {
		t.Error("Expected response.created event")
	}
}

func TestStreamWriter_AutoSequenceNumber(t *testing.T) {
	var buf bytes.Buffer
	flusher := &mockFlusher{}
	writer := NewStreamWriter(&buf, flusher)

	// Create event without sequence number
	event := &BaseStreamingEvent{Type: "test.event"}
	writer.WriteEvent(event)

	// Event should have sequence number set
	if event.SequenceNumber != 1 {
		t.Errorf("Expected sequence number 1, got %d", event.SequenceNumber)
	}

	// Next event should have sequence number 2
	event2 := &BaseStreamingEvent{Type: "test.event"}
	writer.WriteEvent(event2)

	if event2.SequenceNumber != 2 {
		t.Errorf("Expected sequence number 2, got %d", event2.SequenceNumber)
	}
}

func TestStreamWriter_WriteRaw(t *testing.T) {
	var buf bytes.Buffer
	flusher := &mockFlusher{}
	writer := NewStreamWriter(&buf, flusher)

	rawData := []byte("raw data: test\n")
	err := writer.WriteRaw(rawData)
	if err != nil {
		t.Fatalf("WriteRaw failed: %v", err)
	}

	output := buf.String()
	if output != string(rawData) {
		t.Errorf("Expected %q, got %q", string(rawData), output)
	}
}

func TestStreamWriter_WithNilFlusher(t *testing.T) {
	var buf bytes.Buffer
	writer := NewStreamWriter(&buf, nil)

	event := NewResponseCreatedEvent(1, NewResponse("resp_test", "gpt-4o"))
	err := writer.WriteEvent(event)
	if err != nil {
		t.Fatalf("WriteEvent failed: %v", err)
	}

	// Should still work without flusher
	if buf.Len() == 0 {
		t.Error("Expected output even without flusher")
	}
}

func TestStreamWriter_WithNilWriter(t *testing.T) {
	flusher := &mockFlusher{}
	writer := NewStreamWriter(nil, flusher)

	event := NewResponseCreatedEvent(1, NewResponse("resp_test", "gpt-4o"))

	// Should panic when writing to nil writer
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when writer is nil")
		}
	}()
	writer.WriteEvent(event)
}

// Helper function
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func TestResponseCompletedEvent(t *testing.T) {
	resp := NewResponse("resp_123", "gpt-4o")
	resp.Status = ResponseStatusCompleted
	event := NewResponseCompletedEvent(1, resp)

	if event.GetType() != "response.completed" {
		t.Errorf("Expected type 'response.completed', got '%s'", event.GetType())
	}

	if event.SequenceNumber != 1 {
		t.Errorf("Expected sequence number 1, got %d", event.SequenceNumber)
	}
}

func TestStreamingEvent_SequenceNumber(t *testing.T) {
	event := &BaseStreamingEvent{
		Type:          "test.event",
		SequenceNumber: 0,
	}

	if event.GetType() != "test.event" {
		t.Errorf("Expected type 'test.event', got '%s'", event.GetType())
	}

	event.SetSequenceNumber(42)
	if event.GetSequenceNumber() != 42 {
		t.Errorf("Expected sequence number 42, got %d", event.GetSequenceNumber())
	}
}
