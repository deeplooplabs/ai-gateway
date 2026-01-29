package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deeplooplabs/ai-gateway/openresponses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// OpenResponses Tests
// ========================================

// TestE2E_OpenResponses_Basic tests basic non-streaming OpenResponses request
func TestE2E_OpenResponses_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Create OpenResponses request
	reqBody := openresponses.CreateRequest{
		Model: "gpt-4",
		Input: "Hello, how are you?",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Send request
	resp, err := http.Post(
		env.Server.URL+"/v1/responses",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orResp openresponses.Response
	err = json.NewDecoder(resp.Body).Decode(&orResp)
	require.NoError(t, err)

	// Validate OpenResponses format
	assert.NotEmpty(t, orResp.ID, "response should have ID")
	assert.Equal(t, "response", orResp.Object)
	assert.Equal(t, openresponses.ResponseStatusCompleted, orResp.Status)
	assert.Equal(t, "gpt-4", orResp.Model)
	assert.NotEmpty(t, orResp.Output, "response should have output")
	assert.NotNil(t, orResp.Usage, "response should have usage")
	assert.NotNil(t, orResp.CompletedAt, "completed response should have CompletedAt")
}

// TestE2E_OpenResponses_Streaming tests streaming OpenResponses request
func TestE2E_OpenResponses_Streaming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Create streaming request
	stream := true
	reqBody := openresponses.CreateRequest{
		Model:  "gpt-4",
		Input:  "Tell me a story",
		Stream: &stream,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", env.Server.URL+"/v1/responses", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify streaming response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Parse SSE events
	events := parseSSEEvents(t, resp.Body)

	// Verify we received events
	require.NotEmpty(t, events, "should receive at least one event")

	// Track event types we've seen
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	// Verify key events are present
	assert.True(t, eventTypes["response.created"], "should receive response.created event")
	assert.True(t, eventTypes["response.in_progress"], "should receive response.in_progress event")

	// Verify [DONE] marker is present
	foundDone := false
	for _, event := range events {
		if event.IsDone {
			foundDone = true
			break
		}
	}
	assert.True(t, foundDone, "should receive [DONE] marker")
}

// TestE2E_OpenResponses_WithTools tests OpenResponses request with tools
func TestE2E_OpenResponses_WithTools(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Create request with tools
	reqBody := openresponses.CreateRequest{
		Model: "gpt-4",
		Input: "What's the weather like?",
		Tools: []openresponses.Tool{
			openresponses.FunctionTool{
				Type:        "function",
				Name:        "get_weather",
				Description: "Get the current weather in a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city name",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Send request
	resp, err := http.Post(
		env.Server.URL+"/v1/responses",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orResp openresponses.Response
	err = json.NewDecoder(resp.Body).Decode(&orResp)
	require.NoError(t, err)

	assert.NotEmpty(t, orResp.ID)
	assert.Equal(t, "response", orResp.Object)
}

// TestE2E_OpenResponses_ErrorResponse tests OpenResponses error format
func TestE2E_OpenResponses_ErrorResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Create request with invalid model (not registered)
	reqBody := openresponses.CreateRequest{
		Model: "nonexistent-model",
		Input: "Hello",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Send request
	resp, err := http.Post(
		env.Server.URL+"/v1/responses",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify error response
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)

	// Verify OpenResponses error format
	assert.Contains(t, errorResp, "error")
	errorObj := errorResp["error"].(map[string]interface{})
	assert.Contains(t, errorObj, "type")
	assert.Contains(t, errorObj, "message")
	assert.Equal(t, "not_found", errorObj["type"])
}

// ========================================
// Helper Functions
// ========================================

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Event string
	Data  string
	Type  string
	IsDone bool
}

// parseSSEEvents parses SSE stream into events
func parseSSEEvents(t *testing.T, r io.Reader) []SSEEvent {
	t.Helper()

	var events []SSEEvent
	scanner := bufio.NewScanner(r)

	var currentEvent SSEEvent
	for scanner.Scan() {
		line := scanner.Text()

		// Handle [DONE] marker
		if strings.TrimSpace(line) == "data: [DONE]" {
			events = append(events, SSEEvent{IsDone: true})
			continue
		}

		// Empty line indicates end of event
		if line == "" {
			if currentEvent.Event != "" || currentEvent.Data != "" {
				// Try to extract type from data JSON
				if currentEvent.Data != "" {
					var eventData map[string]interface{}
					if err := json.Unmarshal([]byte(currentEvent.Data), &eventData); err == nil {
						if eventType, ok := eventData["type"].(string); ok {
							currentEvent.Type = eventType
						}
					}
				}
				events = append(events, currentEvent)
				currentEvent = SSEEvent{}
			}
			continue
		}

		// Parse event field
		if strings.HasPrefix(line, "event: ") {
			currentEvent.Event = strings.TrimPrefix(line, "event: ")
			continue
		}

		// Parse data field
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if currentEvent.Data == "" {
				currentEvent.Data = data
			} else {
				currentEvent.Data += "\n" + data
			}
			continue
		}
	}

	require.NoError(t, scanner.Err(), "scanner should not error")

	return events
}
