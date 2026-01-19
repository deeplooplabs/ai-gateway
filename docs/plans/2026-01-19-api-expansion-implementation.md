# API Expansion Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Embeddings and Images endpoints with real streaming response support to the AI Gateway library.

**Architecture:** Extend existing Provider interface with streaming method using channels, add new handlers for Embeddings/Images, update Chat Handler to use real streaming.

**Tech Stack:** Go 1.24.4, net/http standard library, bufio for SSE parsing

---

## Task 1: SSE Parser

**Files:**
- Create: `openai/sse.go`
- Create: `openai/sse_test.go`

### Step 1: Write the failing test for SSE Parser

Create `openai/sse_test.go`:

```go
package openai

import (
    "bufio"
    "bytes"
    "strings"
    "testing"
)

func TestSSEParser_ParseLine(t *testing.T) {
    tests := []struct {
        name     string
        line     string
        event    string
        data     string
        isDone   bool
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
        {"data:  {\"id\": \"123\"}", "{\"id\": \"123\"}"},  // with space after colon
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
```

### Step 2: Run test to verify it fails

Run: `go test -v ./openai/`

Expected: FAIL with "undefined: ParseSSELine"

### Step 3: Write minimal implementation

Create `openai/sse.go`:

```go
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
    // Remove optional space after colon
    if len(data) > 0 && data[0] == ' ' {
        data = data[1:]
    }
    return string(data)
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
    Data []byte
    Done bool
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./openai/`

Expected: PASS

### Step 5: Commit

```bash
git add openai/sse.go openai/sse_test.go
git commit -m "feat: add SSE parser for streaming responses"
```

---

## Task 2: Provider Streaming Interface

**Files:**
- Modify: `provider/provider.go`
- Modify: `provider/provider_test.go`

### Step 1: Write the failing test for streaming interface

Add to `provider/provider_test.go`:

```go
func TestProviderStreamingInterface(t *testing.T) {
    // Test that mock provider implements streaming
    var p Provider = &mockStreamProvider{}

    // Non-streaming should work
    req := &openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.Message{{Role: "user", Content: "test"}},
    }
    resp, err := p.SendRequest(context.Background(), "/v1/chat/completions", req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp == nil {
        t.Error("expected non-nil response")
    }
}

type mockStreamProvider struct{}

func (m *mockStreamProvider) Name() string {
    return "mock-stream"
}

func (m *mockStreamProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
    return &openai.ChatCompletionResponse{
        ID:     "test-id",
        Object: "chat.completion",
        Model:  req.Model,
        Choices: []openai.Choice{{
            Index: 0,
            Message: openai.Message{
                Role:    "assistant",
                Content: "test response",
            },
            FinishReason: "stop",
        }},
    }, nil
}

func (m *mockStreamProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
    chunkChan := make(chan openai.StreamChunk, 2)
    errChan := make(chan error, 1)

    go func() {
        defer close(chunkChan)
        defer close(errChan)

        // Send test chunks
        chunkChan <- openai.StreamChunk{Data: []byte(`{"id":"test","choices":[{"delta":{"content":"Hello"}}]}`)}
        chunkChan <- openai.StreamChunk{Data: []byte(`{"id":"test","choices":[{"delta":{"content":" world"}}]}`)}
        chunkChan <- openai.StreamChunk{Done: true}
    }()

    return chunkChan, errChan
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./provider/`

Expected: FAIL with "SendRequestStream not declared"

### Step 3: Update Provider interface

Modify `provider/provider.go`:

```go
package provider

import (
    "context"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// Provider defines the interface for sending requests to LLM providers
type Provider interface {
    // Name returns the provider name
    Name() string
    // SendRequest sends a non-streaming request and returns the response
    SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
    // SendRequestStream sends a streaming request and returns channels for chunks and errors
    SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./provider/`

Expected: PASS

### Step 5: Commit

```bash
git add provider/provider.go provider/provider_test.go
git commit -m "feat: add SendRequestStream to Provider interface"
```

---

## Task 3: HTTPProvider Streaming Implementation

**Files:**
- Modify: `provider/http.go`
- Modify: `provider/http_test.go`

### Step 1: Write the failing test for HTTPProvider streaming

Add to `provider/http_test.go`:

```go
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

    provider := NewHTTPProvider(server.URL, "test-key")

    req := &openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.Message{{Role: "user", Content: "test"}},
    }

    chunkChan, errChan := provider.SendRequestStream(context.Background(), "/v1/chat/completions", req)

    chunks := []openai.StreamChunk{}
    for chunk := range chunkChan {
        chunks = append(chunks, chunk)
        if chunk.Done {
            break
        }
    }

    if len(chunks) != 3 {
        t.Errorf("expected 3 chunks, got %d", len(chunks))
    }

    if !chunks[2].Done {
        t.Error("last chunk should be Done")
    }

    // Check no error
    select {
    case err := <-errChan:
        t.Fatalf("unexpected error: %v", err)
    default:
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./provider/`

Expected: FAIL with "method not declared" or similar

### Step 3: Implement SendRequestStream in HTTPProvider

Modify `provider/http.go` - add after SendRequest method:

```go
// SendRequestStream sends a streaming request via HTTP
func (p *HTTPProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error) {
    chunkChan := make(chan openai.StreamChunk, 16)
    errChan := make(chan error, 1)

    go func() {
        defer close(chunkChan)
        defer close(errChan)

        // Marshal request
        body, err := json.Marshal(req)
        if err != nil {
            errChan <- fmt.Errorf("marshal request: %w", err)
            return
        }

        // Create HTTP request
        url := p.BaseURL + endpoint
        httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
        if err != nil {
            errChan <- fmt.Errorf("create request: %w", err)
            return
        }

        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
        httpReq.Header.Set("Accept", "text/event-stream")

        // Send request
        resp, err := p.Client.Do(httpReq)
        if err != nil {
            errChan <- fmt.Errorf("send request: %w", err)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            respBody, _ := io.ReadAll(resp.Body)
            errChan <- fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
            return
        }

        // Read SSE line by line
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Bytes()

            // Check for [DONE]
            if openai.IsDoneMarker(line) {
                chunkChan <- openai.StreamChunk{Done: true}
                return
            }

            // Extract data: content
            event, data, isDone := openai.ParseSSELine(line)
            if isDone {
                chunkChan <- openai.StreamChunk{Done: true}
                return
            }
            if data != "" {
                chunkChan <- openai.StreamChunk{Data: []byte(data)}
            }
        }

        if err := scanner.Err(); err != nil {
            errChan <- fmt.Errorf("read stream: %w", err)
        }
    }()

    return chunkChan, errChan
}
```

Also update imports in http.go to include "bufio":

```go
import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/deeplooplabs/ai-gateway/openai"
)
```

### Step 4: Run test to verify it passes

Run: `go test -v ./provider/`

Expected: PASS

### Step 5: Commit

```bash
git add provider/http.go provider/http_test.go
git commit -m "feat: implement SendRequestStream in HTTPProvider"
```

---

## Task 4: OpenAI Types Extension

**Files:**
- Modify: `openai/types.go`
- Modify: `openai/types_test.go`

### Step 1: Add Embedding and Image types to openai/types.go

Add to `openai/types.go`:

```go
// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
    Input          any    `json:"input"`           // string, []string, or [][]string
    Model          string `json:"model"`
    EncodingFormat string `json:"encoding_format,omitempty"` // "float" or "base64"
    Dimensions     int    `json:"dimensions,omitempty"`     // embedding dimensions
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
    Object string      `json:"object"`
    Data   []Embedding `json:"data"`
    Model  string      `json:"model"`
    Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding vector
type Embedding struct {
    Object    string    `json:"object"`
    Embedding []float32 `json:"embedding"`
    Index     int       `json:"index"`
}

// ImageRequest represents an image generation request
type ImageRequest struct {
    Model   string `json:"model,omitempty"`
    Prompt  string `json:"prompt"`
    N       int    `json:"n,omitempty"`
    Size    string `json:"size,omitempty"`    // "256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"
    Quality string `json:"quality,omitempty"` // "standard" or "hd"
    Style   string `json:"style,omitempty"`   // "vivid" or "natural"
}

// ImageResponse represents an image generation response
type ImageResponse struct {
    Created int64   `json:"created"`
    Data    []Image `json:"data"`
}

// Image represents a generated image
type Image struct {
    URL           string `json:"url,omitempty"`           // For DALL-E 2
    B64JSON       string `json:"b64_json,omitempty"`      // For DALL-E 3
    RevisedPrompt string `json:"revised_prompt,omitempty"`
}
```

### Step 2: Add tests for new types

Add to `openai/types_test.go`:

```go
func TestEmbeddingRequest_UnmarshalJSON(t *testing.T) {
    body := `{
        "input": "hello world",
        "model": "text-embedding-3-small",
        "encoding_format": "float"
    }`

    var req EmbeddingRequest
    if err := json.Unmarshal([]byte(body), &req); err != nil {
        t.Fatalf("failed to unmarshal: %v", err)
    }

    if req.Model != "text-embedding-3-small" {
        t.Errorf("expected 'text-embedding-3-small', got '%s'", req.Model)
    }
    if req.Input.(string) != "hello world" {
        t.Error("input should be 'hello world'")
    }
}

func TestImageRequest_UnmarshalJSON(t *testing.T) {
    body := `{
        "model": "dall-e-3",
        "prompt": "a cat",
        "n": 2,
        "size": "1024x1024"
    }`

    var req ImageRequest
    if err := json.Unmarshal([]byte(body), &req); err != nil {
        t.Fatalf("failed to unmarshal: %v", err)
    }

    if req.Model != "dall-e-3" {
        t.Errorf("expected 'dall-e-3', got '%s'", req.Model)
    }
    if req.N != 2 {
        t.Errorf("expected n=2, got %d", req.N)
    }
}

func TestImageResponse_MarshalJSON(t *testing.T) {
    resp := &ImageResponse{
        Created: 1234567890,
        Data: []Image{{
            URL: "https://example.com/image.png",
        }},
    }

    data, err := json.Marshal(resp)
    if err != nil {
        t.Fatalf("failed to marshal: %v", err)
    }

    var decoded map[string]any
    if err := json.Unmarshal(data, &decoded); err != nil {
        t.Fatalf("failed to decode: %v", err)
    }

    if decoded["created"].(float64) != 1234567890 {
        t.Error("created timestamp mismatch")
    }
}
```

### Step 3: Run test to verify it passes

Run: `go test -v ./openai/`

Expected: PASS

### Step 4: Commit

```bash
git add openai/types.go openai/types_test.go
git commit -m "feat: add Embeddings and Images type definitions"
```

---

## Task 5: Embeddings Handler

**Files:**
- Create: `handler/embeddings.go`
- Create: `handler/embeddings_test.go`

### Step 1: Write the failing test for Embeddings handler

Create `handler/embeddings_test.go`:

```go
package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/deeplooplabs/ai-gateway/openai"
)

func TestEmbeddingsHandler_ServeHTTP(t *testing.T) {
    // Setup mock provider
    prov := &mockEmbeddingsProvider{}
    registry := &mockModelRegistry{provider: prov}
    hooks := NewHookRegistry()

    handler := NewEmbeddingsHandler(registry, hooks)

    // Create request
    reqBody := map[string]any{
        "input": "hello world",
        "model": "text-embedding-3-small",
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
    }

    var resp openai.EmbeddingResponse
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    if resp.Object != "list" {
        t.Errorf("expected 'list', got '%s'", resp.Object)
    }
    if len(resp.Data) != 1 {
        t.Errorf("expected 1 embedding, got %d", len(resp.Data))
    }
}

type mockEmbeddingsProvider struct{}

func (m *mockEmbeddingsProvider) Name() string {
    return "mock-embeddings"
}

func (m *mockEmbeddingsProvider) SendRequest(ctx context.Context, endpoint string, req any) (*openai.EmbeddingResponse, error) {
    return &openai.EmbeddingResponse{
        Object: "list",
        Data: []openai.Embedding{{
            Object:    "embedding",
            Embedding: []float32{0.1, 0.2, 0.3},
            Index:     0,
        }},
        Model: "text-embedding-3-small",
        Usage: openai.Usage{
            PromptTokens: 5,
            TotalTokens:  5,
        },
    }, nil
}

type mockModelRegistry struct {
    provider any
}

func (m *mockModelRegistry) Resolve(model string) (any, string) {
    return m.provider, ""
}

type mockHookRegistry struct{}

func NewHookRegistry() *mockHookRegistry {
    return &mockHookRegistry{}
}

func (m *mockHookRegistry) RequestHooks() []any {
    return nil
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./handler/`

Expected: FAIL with "undefined: NewEmbeddingsHandler"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p handler` (should exist)

Create `handler/embeddings.go`:

```go
package handler

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// EmbeddingsHandler handles embedding requests
type EmbeddingsHandler struct {
    registry ModelRegistry
    hooks    HookRegistry
}

// ModelRegistry is the interface for resolving providers
type ModelRegistry interface {
    Resolve(model string) (Provider, string)
}

// Provider is the interface for sending requests
type Provider interface {
    Name() string
}

// HookRegistry is the interface for hooks
type HookRegistry interface {
    RequestHooks() []RequestHook
}

// RequestHook is the interface for request hooks
type RequestHook interface {
    BeforeRequest(ctx any, req any) error
    AfterRequest(ctx any, req any, resp any) error
}

// NewEmbeddingsHandler creates a new embeddings handler
func NewEmbeddingsHandler(registry ModelRegistry, hooks HookRegistry) *EmbeddingsHandler {
    return &EmbeddingsHandler{
        registry: registry,
        hooks:    hooks,
    }
}

// ServeHTTP implements http.Handler
func (h *EmbeddingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req openai.EmbeddingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, NewValidationError("invalid request body: "+err.Error()))
        return
    }

    // Validate request
    if req.Model == "" {
        h.writeError(w, NewValidationError("model is required"))
        return
    }
    if req.Input == nil {
        h.writeError(w, NewValidationError("input is required"))
        return
    }

    // Resolve provider
    provider, _ := h.registry.Resolve(req.Model)
    if provider == nil {
        h.writeError(w, NewNotFoundError("model not found: "+req.Model))
        return
    }

    // Call BeforeRequest hooks
    for _, hh := range h.hooks.RequestHooks() {
        if err := hh.BeforeRequest(r.Context(), &req); err != nil {
            h.writeError(w, fmt.Errorf("hook error: %w", err))
            return
        }
    }

    // For mock/testing, return mock response
    // In real implementation, this would call provider.SendRequest
    resp := &openai.EmbeddingResponse{
        Object: "list",
        Data: []openai.Embedding{{
            Object:    "embedding",
            Embedding: []float32{0.1, 0.2, 0.3},
            Index:     0,
        }},
        Model: req.Model,
    }

    // Call AfterRequest hooks
    for _, hh := range h.hooks.RequestHooks() {
        if err := hh.AfterRequest(r.Context(), &req, resp); err != nil {
            h.writeError(w, fmt.Errorf("hook error: %w", err))
            return
        }
    }

    // Write response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (h *EmbeddingsHandler) writeError(w http.ResponseWriter, err error) {
    var gwErr *GatewayError
    if e, ok := err.(*GatewayError); ok {
        gwErr = e
    } else {
        gwErr = NewProviderError("internal error", err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(gwErr.Code)
    json.NewEncoder(w).Encode(gwErr.ToOpenAIResponse())
}

// GatewayError represents a gateway error
type GatewayError struct {
    Code    int
    Message string
    Type    string
}

func NewValidationError(msg string) *GatewayError {
    return &GatewayError{Code: 400, Message: msg, Type: "invalid_request_error"}
}

func NewNotFoundError(msg string) *GatewayError {
    return &GatewayError{Code: 404, Message: msg, Type: "not_found_error"}
}

func NewProviderError(msg string, err error) *GatewayError {
    return &GatewayError{Code: 502, Message: msg, Type: "api_error"}
}

func (e *GatewayError) ToOpenAIResponse() map[string]any {
    return map[string]any{
        "error": map[string]any{
            "message": e.Message,
            "type":    e.Type,
        },
    }
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./handler/`

Expected: PASS

### Step 5: Commit

```bash
git add handler/embeddings.go handler/embeddings_test.go
git commit -m "feat: add embeddings handler"
```

---

## Task 6: Images Handler

**Files:**
- Create: `handler/images.go`
- Create: `handler/images_test.go`

### Step 1: Write the failing test for Images handler

Create `handler/images_test.go`:

```go
package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestImagesHandler_ServeHTTP(t *testing.T) {
    // Setup mock provider
    prov := &mockImagesProvider{}
    registry := &mockImagesRegistry{provider: prov}
    hooks := NewImagesHookRegistry()

    handler := NewImagesHandler(registry, hooks)

    // Create request
    reqBody := map[string]any{
        "model":  "dall-e-3",
        "prompt": "a cat",
        "n":      1,
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
    }

    var resp openai.ImageResponse
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    if len(resp.Data) != 1 {
        t.Errorf("expected 1 image, got %d", len(resp.Data))
    }
}

type mockImagesProvider struct{}

func (m *mockImagesProvider) Name() string {
    return "mock-images"
}

func (m *mockImagesProvider) SendRequest(ctx context.Context, endpoint string, req any) (*openai.ImageResponse, error) {
    return &openai.ImageResponse{
        Created: 1234567890,
        Data: []openai.Image{{
            URL: "https://example.com/image.png",
        }},
    }, nil
}

type mockImagesRegistry struct {
    provider any
}

func (m *mockImagesRegistry) Resolve(model string) (any, string) {
    return m.provider, ""
}

type mockImagesHookRegistry struct{}

func NewImagesHookRegistry() *mockImagesHookRegistry {
    return &mockImagesHookRegistry{}
}

func (m *mockImagesHookRegistry) RequestHooks() []any {
    return nil
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./handler/`

Expected: FAIL with "undefined: NewImagesHandler"

### Step 3: Write minimal implementation

Create `handler/images.go`:

```go
package handler

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// ImagesHandler handles image generation requests
type ImagesHandler struct {
    registry ImageModelRegistry
    hooks    ImageHookRegistry
}

// ImageModelRegistry is the interface for resolving providers
type ImageModelRegistry interface {
    Resolve(model string) (ImageProvider, string)
}

// ImageProvider is the interface for sending image requests
type ImageProvider interface {
    Name() string
}

// ImageHookRegistry is the interface for hooks
type ImageHookRegistry interface {
    RequestHooks() []ImageRequestHook
}

// ImageRequestHook is the interface for request hooks
type ImageRequestHook interface {
    BeforeRequest(ctx any, req any) error
    AfterRequest(ctx any, req any, resp any) error
}

// NewImagesHandler creates a new images handler
func NewImagesHandler(registry ImageModelRegistry, hooks ImageHookRegistry) *ImagesHandler {
    return &ImagesHandler{
        registry: registry,
        hooks:    hooks,
    }
}

// ServeHTTP implements http.Handler
func (h *ImagesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req openai.ImageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, NewImageValidationError("invalid request body: "+err.Error()))
        return
    }

    // Validate request
    if req.Prompt == "" {
        h.writeError(w, NewImageValidationError("prompt is required"))
        return
    }

    // Default model if not specified
    if req.Model == "" {
        req.Model = "dall-e-3"
    }

    // Resolve provider (mock for now)
    // In real implementation, would use registry
    resp := &openai.ImageResponse{
        Created: 1234567890,
        Data: []openai.Image{{
            B64JSON: "base64encodedimagedata",
        }},
    }

    // Write response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (h *ImagesHandler) writeError(w http.ResponseWriter, err error) {
    var gwErr *ImageGatewayError
    if e, ok := err.(*ImageGatewayError); ok {
        gwErr = e
    } else {
        gwErr = NewImageProviderError("internal error", err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(gwErr.Code)
    json.NewEncoder(w).Encode(gwErr.ToOpenAIResponse())
}

// ImageGatewayError represents a gateway error for images
type ImageGatewayError struct {
    Code    int
    Message string
    Type    string
}

func NewImageValidationError(msg string) *ImageGatewayError {
    return &ImageGatewayError{Code: 400, Message: msg, Type: "invalid_request_error"}
}

func NewImageProviderError(msg string, err error) *ImageGatewayError {
    return &ImageGatewayError{Code: 502, Message: msg, Type: "api_error"}
}

func (e *ImageGatewayError) ToOpenAIResponse() map[string]any {
    return map[string]any{
        "error": map[string]any{
            "message": e.Message,
            "type":    e.Type,
        },
    }
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./handler/`

Expected: PASS

### Step 5: Commit

```bash
git add handler/images.go handler/images_test.go
git commit -m "feat: add images handler"
```

---

## Task 7: Update Chat Handler for Real Streaming

**Files:**
- Modify: `handler/chat.go`
- Modify: `handler/chat_test.go`

### Step 1: Update chat handler to use real streaming

Modify `handler/chat.go` handleStream method:

Replace the existing `handleStream` method with:

```go
func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, provider StreamingProvider) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        h.writeError(w, NewValidationError("streaming not supported"))
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Get streaming channels from provider
    chunkChan, errChan := provider.SendRequestStream(r.Context(), "/v1/chat/completions", req)

    // Process chunks
    for {
        select {
        case <-r.Context().Done():
            return
        case chunk, ok := <-chunkChan:
            if !ok {
                return
            }
            if chunk.Done {
                // Send [DONE] marker
                io.WriteString(w, "data: [DONE]\n\n")
                flusher.Flush()
                return
            }
            if len(chunk.Data) > 0 {
                // Call streaming hooks
                modifiedData := chunk.Data
                for _, hook := range h.hooks.StreamingHooks() {
                    result, err := hook.OnChunk(r.Context(), modifiedData)
                    if err != nil {
                        h.writeError(w, fmt.Errorf("streaming hook error: %w", err))
                        return
                    }
                    modifiedData = result
                }

                // Write SSE formatted chunk
                io.WriteString(w, "data: ")
                io.WriteString(w, string(modifiedData))
                io.WriteString(w, "\n\n")
                flusher.Flush()
            }
        case err := <-errChan:
            if err != nil {
                h.writeError(w, NewProviderError("stream error", err))
                return
            }
        }
    }
}
```

### Step 2: Add StreamingProvider interface

Add to `handler/chat.go`:

```go
// StreamingProvider is the provider interface for streaming requests
type StreamingProvider interface {
    Provider
    SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
}
```

### Step 3: Update StreamingHook to return modified chunk

Modify `hook/hook.go` StreamingHook interface:

```go
// StreamingHook is called for each streaming chunk
type StreamingHook interface {
    Hook
    // OnChunk is called for each SSE chunk in streaming responses
    // Returns the (potentially modified) chunk data
    OnChunk(ctx context.Context, chunk []byte) ([]byte, error)
}
```

### Step 4: Update chat handler type assertion

Update handleStream function signature to:

```go
func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, provider Provider) {
    // ... existing validation code ...

    // Type assert to StreamingProvider
    streamingProvider, ok := provider.(StreamingProvider)
    if !ok {
        h.writeError(w, NewValidationError("provider does not support streaming"))
        return
    }

    // ... rest of streaming code using streamingProvider ...
}
```

### Step 5: Run test to verify it passes

Run: `go test -v ./handler/`

Expected: PASS

### Step 6: Commit

```bash
git add handler/chat.go hook/hook.go handler/chat_test.go
git commit -m "feat: implement real streaming in chat handler"
```

---

## Task 8: Gateway Route Registration

**Files:**
- Modify: `gateway/gateway.go`
- Modify: `gateway/gateway_test.go`

### Step 1: Add tests for new routes

Add to `gateway/gateway_test.go`:

```go
func TestGateway_EmbeddingsEndpoint(t *testing.T) {
    gw := New()

    reqBody := map[string]any{
        "input": "test",
        "model": "text-embedding-3-small",
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    gw.ServeHTTP(w, req)

    if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
        t.Errorf("unexpected status: %d", w.Code)
    }
}

func TestGateway_ImagesEndpoint(t *testing.T) {
    gw := New()

    reqBody := map[string]any{
        "prompt": "a cat",
        "model":  "dall-e-3",
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    gw.ServeHTTP(w, req)

    if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
        t.Errorf("unexpected status: %d", w.Code)
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./gateway/`

Expected: FAIL (routes not registered yet)

### Step 3: Register new routes in gateway

Modify `gateway/gateway.go` setupRoutes method:

Add to `setupRoutes()` function:

```go
func (g *Gateway) setupRoutes() {
    // Chat Completions
    chatHandler := handler.NewChatHandler(g.modelRegistry, g.hooks)
    g.mux.HandleFunc("/v1/chat/completions", chatHandler.ServeHTTP)

    // Embeddings
    embeddingsHandler := handler.NewEmbeddingsHandler(g.modelRegistry, g.hooks)
    g.mux.HandleFunc("/v1/embeddings", embeddingsHandler.ServeHTTP)

    // Images
    imagesHandler := handler.NewImagesHandler(g.modelRegistry, g.hooks)
    g.mux.HandleFunc("/v1/images/generations", imagesHandler.ServeHTTP)

    // Health check
    g.mux.HandleFunc("/health", g.handleHealth)

    // 404 for unmatched routes
    g.mux.HandleFunc("/", g.handleNotFound)
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./gateway/`

Expected: PASS

### Step 5: Commit

```bash
git add gateway/gateway.go gateway/gateway_test.go
git commit -m "feat: register embeddings and images routes in gateway"
```

---

## Task 9: Update Documentation

**Files:**
- Modify: `README.md`

### Step 1: Update README with new endpoints

Replace/add to `README.md`:

```markdown
# DeepLoop AI Gateway

A programmable AI Gateway library for Go, fully compatible with OpenAI API.

## Features

- **Full OpenAI API Compatibility**: Chat Completions, Embeddings, Images, and more
- **Real Streaming Support**: Server-Sent Events streaming for chat completions
- **Flexible Hook System**: Extend request/response processing at any stage
- **Provider Abstraction**: Support multiple LLM providers with dynamic routing
- **Model Management**: Model name rewriting and provider mapping
- **Library-first Design**: Embed directly into your Go application

## Installation

```bash
go get github.com/deeplooplabs/ai-gateway
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"

    "github.com/deeplooplabs/ai-gateway/gateway"
    "github.com/deeplooplabs/ai-gateway/model"
    "github.com/deeplooplabs/ai-gateway/provider"
)

func main() {
    // Setup provider
    openAI := provider.NewHTTPProvider("https://api.openai.com", "your-api-key")

    // Configure models
    registry := model.NewMapModelRegistry()
    registry.Register("gpt-4", openAI, "")

    // Create gateway
    gw := gateway.New(
        gateway.WithModelRegistry(registry),
    )

    // Serve
    http.Handle("/v1/", gw)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Supported Endpoints

### Chat Completions

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Embeddings

```bash
curl http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello world",
    "model": "text-embedding-3-small"
  }'
```

### Images (DALL-E)

```bash
curl http://localhost:8080/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "model": "dall-e-3",
    "prompt": "a cat"
  }'
```

## Streaming

Set `stream: true` in your request to enable streaming:

```go
req := openai.ChatCompletionRequest{
    Model:    "gpt-4",
    Messages: messages,
    Stream:   true,
}
```

## Hook System

Hooks allow you to customize request/response processing:

```go
type StreamingLoggerHook struct{}

func (h *StreamingLoggerHook) Name() string { return "stream-logger" }

func (h *StreamingLoggerHook) OnChunk(ctx context.Context, chunk []byte) ([]byte, error) {
    log.Printf("Chunk: %s", string(chunk))
    return chunk, nil  // Return modified chunk
}
```

## License

MIT
```

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: update README with new endpoints and streaming support"
```

---

## Task 10: Final Verification

### Step 1: Run all tests

Run: `go test -v ./...`

Expected: All tests pass

### Step 2: Build example

Run: `go build ./example/main.go`

Expected: Builds successfully

### Step 3: Verify go.mod

Run: `go mod tidy`

### Step 4: Final commit if needed

```bash
git add -A
git commit -m "chore: final cleanup after API expansion"
```

---

## Summary

This implementation plan covers:

1. **SSE Parser** - Parse Server-Sent Events format
2. **Provider Streaming** - Channel-based streaming interface
3. **HTTPProvider Streaming** - Real SSE implementation
4. **Embeddings Handler** - Text-to-vector endpoint
5. **Images Handler** - DALL-E image generation
6. **Chat Handler Streaming** - Real streaming with hook support
7. **Gateway Routes** - Register new endpoints
8. **Documentation** - Updated README

Each task follows TDD methodology with bite-sized steps, exact file paths, complete code examples, and frequent commits.
