package openresponses

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StreamWriter writes OpenResponses streaming events in SSE format
type StreamWriter struct {
	writer   io.Writer
	flusher  http.Flusher
	sequence int
}

// NewStreamWriter creates a new StreamWriter
func NewStreamWriter(w io.Writer, flusher http.Flusher) *StreamWriter {
	return &StreamWriter{
		writer:  w,
		flusher: flusher,
	}
}

// WriteEvent writes a single streaming event
func (w *StreamWriter) WriteEvent(event StreamingEvent) error {
	// Set sequence number if not already set
	if event.GetSequenceNumber() == 0 {
		event.SetSequenceNumber(w.NextSequence())
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Write SSE format: event: <type>\ndata: <json>\n\n
	eventType := event.GetType()
	if eventType != "" {
		if _, err := fmt.Fprintf(w.writer, "event: %s\n", eventType); err != nil {
			return fmt.Errorf("write event type: %w", err)
		}
	}

	if _, err := fmt.Fprintf(w.writer, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("write event data: %w", err)
	}

	// Flush to ensure immediate delivery
	if w.flusher != nil {
		w.flusher.Flush()
	}

	return nil
}

// WriteDone writes the [DONE] marker to end the stream
func (w *StreamWriter) WriteDone() error {
	if _, err := fmt.Fprint(w.writer, "data: [DONE]\n\n"); err != nil {
		return fmt.Errorf("write done marker: %w", err)
	}
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}

// WriteError writes an error event and terminates the stream
func (w *StreamWriter) WriteError(err *Error) error {
	seq := w.NextSequence()
	event := NewErrorStreamingEvent(seq, err)
	if writeErr := w.WriteEvent(event); writeErr != nil {
		return writeErr
	}
	return w.WriteDone()
}

// NextSequence returns the next sequence number
func (w *StreamWriter) NextSequence() int {
	w.sequence++
	return w.sequence
}

// WriteRaw writes raw SSE data (for compatibility with existing providers)
func (w *StreamWriter) WriteRaw(data []byte) error {
	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("write raw data: %w", err)
	}
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}
