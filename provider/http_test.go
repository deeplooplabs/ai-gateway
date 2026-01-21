package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	openai2 "github.com/deeplooplabs/ai-gateway/provider/openai"
)

func TestHTTPProvider_SendRequestStream(t *testing.T) {
	// Setup test server
	sseResponse := `event: message
data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"content":"Hello"}}],"object":"chat.completion.chunk"}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"content":" world"}}],"object":"chat.completion.chunk"}

data: [DONE]
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte(sseResponse))
	}))
	defer server.Close()

	provider := NewHTTPProviderWithBaseURL(server.URL, "test-key")

	req := NewChatCompletionsRequest("gpt-4", []openai2.Message{{Role: "user", Content: "test"}})
	req.Stream = true
	req.Endpoint = "/v1/chat/completions"

	resp, err := provider.SendRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Close()

	if !resp.Stream {
		t.Fatal("expected streaming response")
	}

	chunks := []*Chunk{}
	for chunk := range resp.Chunks {
		chunks = append(chunks, chunk)
		if chunk.Done {
			break
		}
	}

	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}

	if len(chunks) >= 3 && !chunks[2].Done {
		t.Error("last chunk should be Done")
	}

	// Check no error - the channel should be closed with no error sent
	// When closed, reading returns nil (zero value for error)
	select {
	case err := <-resp.Errors:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
		// No error
	}
}
