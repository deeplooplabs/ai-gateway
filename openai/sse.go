package openai

import (
	"bytes"
)

// ParseSSELine parses a single SSE line, returning (event, data, isDone)
func ParseSSELine(line []byte) (event, data string, isDone bool) {
	// Skip empty lines
	if len(line) == 0 {
		return "", "", false
	}

	// Skip comment lines (starting with :)
	if line[0] == ':' {
		return "", "", false
	}

	// Check for [DONE] marker
	if bytes.HasPrefix(line, []byte("data: [DONE]")) {
		return "", "", true
	}

	// Parse "event: xxx"
	if bytes.HasPrefix(line, []byte("event:")) {
		return string(bytes.TrimPrefix(line, []byte("event: "))), "", false
	}

	// Parse "data: xxx"
	if bytes.HasPrefix(line, []byte("data:")) {
		return "", ExtractData(line), false
	}

	return "", "", false
}

// IsDoneMarker checks if the line is a [DONE] marker
func IsDoneMarker(line []byte) bool {
	return bytes.HasPrefix(line, []byte("data: [DONE]"))
}

// ExtractData extracts the data portion from a "data: xxx" line
func ExtractData(line []byte) string {
	// Remove "data:" prefix
	data := bytes.TrimPrefix(line, []byte("data:"))
	// Trim leading spaces
	data = bytes.TrimLeft(data, " ")
	return string(data)
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
	Data []byte
	Done bool
}
