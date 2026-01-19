package openai

import (
	"bufio"
	"bytes"
	"io"
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

// SSEScanner scans Server-Sent Events (SSE) streams
type SSEScanner struct {
	scanner *bufio.Scanner
	err     error
	chunk   StreamChunk
}

// NewSSEScanner creates a new SSE scanner
func NewSSEScanner(r io.Reader) *SSEScanner {
	return &SSEScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan advances to the next SSE chunk. Returns false when stream ends.
func (s *SSEScanner) Scan() bool {
	for s.scanner.Scan() {
		line := s.scanner.Bytes()
		_, data, isDone := ParseSSELine(line)

		if isDone {
			s.chunk = StreamChunk{Done: true}
			return true
		}

		if data != "" {
			s.chunk = StreamChunk{Data: []byte(data)}
			return true
		}
	}

	s.err = s.scanner.Err()
	return s.err == nil
}

// Chunk returns the current chunk
func (s *SSEScanner) Chunk() StreamChunk {
	return s.chunk
}

// Err returns any error encountered during scanning
func (s *SSEScanner) Err() error {
	return s.err
}
