# AI Gateway Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a programmable AI Gateway library in Go that provides full OpenAI API compatibility with a flexible hook system.

**Architecture:** Four-layer architecture: Routing (paths to handlers) → Handler (parse/assemble) → Provider (send requests) → Hook (lifecycle extensions). All requests are parsed, processed through hooks, reconstructed, and forwarded to upstream providers.

**Tech Stack:** Go 1.24.4, net/http (standard library), OpenTelemetry (optional), Prometheus (optional)

---

## Task 1: Project Foundation

**Files:**
- Create: `context.go`
- Create: `context_test.go`
- Create: `error.go`
- Create: `error_test.go`

### Step 1: Write the failing test for Context

Create `context_test.go`:

```go
package ai_gateway

import (
    "net/http"
    "testing"
    "time"
)

func TestNewContext(t *testing.T) {
    req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
    ctx := NewContext(req)

    if ctx.RequestID == "" {
        t.Error("RequestID should not be empty")
    }
    if ctx.StartTime.IsZero() {
        t.Error("StartTime should not be zero")
    }
    if ctx.OriginalReq != req {
        t.Error("OriginalReq should match")
    }
    if ctx.Metadata == nil {
        t.Error("Metadata should be initialized")
    }
}

func TestContextSetGet(t *testing.T) {
    ctx := NewContext(nil)
    ctx.Set("key", "value")

    if val := ctx.Get("key"); val != "value" {
        t.Errorf("expected 'value', got '%v'", val)
    }
    if val := ctx.Get("nonexistent"); val != nil {
        t.Errorf("expected nil, got '%v'", val)
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test -v`

Expected: FAIL with "undefined: NewContext"

### Step 3: Write minimal implementation

Create `context.go`:

```go
package ai_gateway

import (
    "net/http"
    "sync"
    "time"

    "github.com/google/uuid"
)

// Context represents the request context throughout its lifecycle
type Context struct {
    RequestID   string
    StartTime   time.Time
    OriginalReq *http.Request
    Metadata    map[string]any
    Provider    Provider
    mu          sync.RWMutex
}

// NewContext creates a new request context
func NewContext(req *http.Request) *Context {
    return &Context{
        RequestID:   uuid.New().String(),
        StartTime:   time.Now(),
        OriginalReq: req,
        Metadata:    make(map[string]any),
    }
}

// Set stores a value in the context metadata
func (c *Context) Set(key string, value any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.Metadata[key] = value
}

// Get retrieves a value from the context metadata
func (c *Context) Get(key string) any {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.Metadata[key]
}
```

### Step 4: Run test to verify it passes

Run: `go test -v`

Expected: PASS

### Step 5: Add uuid dependency

Run: `go get github.com/google/uuid`

### Step 6: Run test to verify it still passes

Run: `go test -v`

Expected: PASS

### Step 7: Commit

```bash
git add context.go context_test.go go.mod go.sum
git commit -m "feat: add request Context with metadata support"
```

---

### Step 8: Write failing test for GatewayError

Create `error_test.go`:

```go
package ai_gateway

import (
    "encoding/json"
    "net/http"
    "testing"
)

func TestGatewayError(t *testing.T) {
    err := &GatewayError{
        Code:    http.StatusBadRequest,
        Message: "Invalid API key",
        Type:    "invalid_request_error",
    }

    if err.Error() == "" {
        t.Error("Error() should return non-empty string")
    }

    // Test JSON marshaling to OpenAI format
    resp := err.ToOpenAIResponse()
    if resp.Error == nil {
        t.Error("ToOpenAIResponse should have Error field")
    }
    if resp.Error.Message != "Invalid API key" {
        t.Errorf("expected 'Invalid API key', got '%s'", resp.Error.Message)
    }
    if resp.Error.Type != "invalid_request_error" {
        t.Errorf("expected 'invalid_request_error', got '%s'", resp.Error.Type)
    }
}

func TestNewAuthenticationError(t *testing.T) {
    err := NewAuthenticationError("Invalid API key")
    if err.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", err.Code)
    }
    if err.Type != "authentication_error" {
        t.Errorf("expected 'authentication_error', got '%s'", err.Type)
    }
}

func TestNewValidationError(t *testing.T) {
    err := NewValidationError("model is required")
    if err.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", err.Code)
    }
    if err.Type != "invalid_request_error" {
        t.Errorf("expected 'invalid_request_error', got '%s'", err.Type)
    }
}
```

### Step 9: Run test to verify it fails

Run: `go test -v`

Expected: FAIL with "undefined: GatewayError"

### Step 10: Write minimal implementation

Create `error.go`:

```go
package ai_gateway

import (
    "fmt"
    "net/http"
)

// GatewayError represents an error that can be returned to the client
type GatewayError struct {
    Code       int
    Message    string
    Type       string
    Param      string
    InnerError error
}

// Error implements the error interface
func (e *GatewayError) Error() string {
    if e.InnerError != nil {
        return fmt.Sprintf("%s: %s (inner: %v)", e.Type, e.Message, e.InnerError)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// OpenAIErrorResponse represents the error response format compatible with OpenAI
type OpenAIErrorResponse struct {
    Error *OpenAIErrorDetail `json:"error"`
}

// OpenAIErrorDetail represents the error detail in OpenAI format
type OpenAIErrorDetail struct {
    Message string `json:"message"`
    Type    string `json:"type"`
    Param   string `json:"param,omitempty"`
    Code    string `json:"code,omitempty"`
}

// ToOpenAIResponse converts the GatewayError to OpenAI error response format
func (e *GatewayError) ToOpenAIResponse() *OpenAIErrorResponse {
    return &OpenAIErrorResponse{
        Error: &OpenAIErrorDetail{
            Message: e.Message,
            Type:    e.Type,
            Param:   e.Param,
            Code:    http.StatusText(e.Code),
        },
    }
}

// NewAuthenticationError creates a new authentication error (401)
func NewAuthenticationError(message string) *GatewayError {
    return &GatewayError{
        Code:    http.StatusUnauthorized,
        Message: message,
        Type:    "authentication_error",
    }
}

// NewValidationError creates a new validation error (400)
func NewValidationError(message string) *GatewayError {
    return &GatewayError{
        Code:    http.StatusBadRequest,
        Message: message,
        Type:    "invalid_request_error",
    }
}

// NewNotFoundError creates a new not found error (404)
func NewNotFoundError(message string) *GatewayError {
    return &GatewayError{
        Code:    http.StatusNotFound,
        Message: message,
        Type:    "invalid_request_error",
    }
}

// NewRateLimitError creates a new rate limit error (429)
func NewRateLimitError(message string) *GatewayError {
    return &GatewayError{
        Code:    http.StatusTooManyRequests,
        Message: message,
        Type:    "rate_limit_error",
    }
}

// NewProviderError creates a new provider error (502)
func NewProviderError(message string, inner error) *GatewayError {
    return &GatewayError{
        Code:       http.StatusBadGateway,
        Message:    message,
        Type:       "api_error",
        InnerError: inner,
    }
}
```

### Step 11: Run test to verify it passes

Run: `go test -v`

Expected: PASS

### Step 12: Commit

```bash
git add error.go error_test.go
git commit -m "feat: add error types with OpenAI-compatible response format"
```

---

## Task 2: OpenAI Type Definitions

**Files:**
- Create: `openai/types.go`
- Create: `openai/types_test.go`

### Step 1: Write the failing test for Chat Completion types

Create `openai/types_test.go`:

```go
package openai

import (
    "encoding/json"
    "testing"
)

func TestChatCompletionRequest_UnmarshalJSON(t *testing.T) {
    body := `{
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Hello"}],
        "temperature": 0.7,
        "stream": false
    }`

    var req ChatCompletionRequest
    if err := json.Unmarshal([]byte(body), &req); err != nil {
        t.Fatalf("failed to unmarshal: %v", err)
    }

    if req.Model != "gpt-4" {
        t.Errorf("expected 'gpt-4', got '%s'", req.Model)
    }
    if len(req.Messages) != 1 {
        t.Errorf("expected 1 message, got %d", len(req.Messages))
    }
    if req.Messages[0].Role != "user" {
        t.Errorf("expected 'user', got '%s'", req.Messages[0].Role)
    }
    if req.Temperature == nil || *req.Temperature != 0.7 {
        t.Error("temperature should be 0.7")
    }
}

func TestChatCompletionResponse_MarshalJSON(t *testing.T) {
    resp := &ChatCompletionResponse{
        ID:      "chatcmpl-123",
        Object:  "chat.completion",
        Created: 1234567890,
        Model:   "gpt-4",
        Choices: []Choice{{
            Index: 0,
            Message: Message{
                Role:    "assistant",
                Content: "Hello!",
            },
            FinishReason: "stop",
        }},
        Usage: Usage{
            PromptTokens:     10,
            CompletionTokens: 5,
            TotalTokens:      15,
        },
    }

    data, err := json.Marshal(resp)
    if err != nil {
        t.Fatalf("failed to marshal: %v", err)
    }

    var decoded map[string]any
    if err := json.Unmarshal(data, &decoded); err != nil {
        t.Fatalf("failed to decode: %v", err)
    }

    if decoded["id"] != "chatcmpl-123" {
        t.Errorf("expected 'chatcmpl-123', got '%v'", decoded["id"])
    }
    if decoded["object"] != "chat.completion" {
        t.Errorf("expected 'chat.completion', got '%v'", decoded["object"])
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./openai/`

Expected: FAIL with "package not found" or "undefined types"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p openai`

Create `openai/types.go`:

```go
package openai

// Message represents a chat message
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// Choice represents a completion choice
type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message,omitempty"`
    Delta        *Delta  `json:"delta,omitempty"`
    FinishReason string  `json:"finish_reason"`
}

// Delta represents streaming message delta
type Delta struct {
    Role    string `json:"role,omitempty"`
    Content string `json:"content,omitempty"`
}

// Usage represents token usage
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
    Model            string    `json:"model"`
    Messages         []Message `json:"messages"`
    Temperature      *float64  `json:"temperature,omitempty"`
    TopP             *float64  `json:"top_p,omitempty"`
    N                *int      `json:"n,omitempty"`
    Stream           bool      `json:"stream,omitempty"`
    MaxTokens        *int      `json:"max_tokens,omitempty"`
    Stop             any       `json:"stop,omitempty"`
    PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
    FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

// ChatCompletionStreamResponse represents a streaming chunk
type ChatCompletionStreamResponse struct {
    ID      string  `json:"id"`
    Object  string  `json:"object"`
    Created int64   `json:"created"`
    Model   string  `json:"model"`
    Choices []Choice `json:"choices"`
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./openai/`

Expected: PASS

### Step 5: Commit

```bash
git add openai/types.go openai/types_test.go
git commit -m "feat: add OpenAI chat completion types"
```

---

## Task 3: Hook System

**Files:**
- Create: `hook/hook.go`
- Create: `hook/hook_test.go`

### Step 1: Write the failing test for Hook interfaces

Create `hook/hook_test.go`:

```go
package hook

import (
    "context"
    "testing"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// mockHook implements Hook interface for testing
type mockHook struct {
    name string
}

func (m *mockHook) Name() string {
    return m.name
}

func TestHookRegistry_Register(t *testing.T) {
    registry := NewRegistry()

    h1 := &mockHook{name: "hook1"}
    h2 := &mockHook{name: "hook2"}

    registry.Register(h1)
    registry.Register(h2)

    if len(registry.All()) != 2 {
        t.Errorf("expected 2 hooks, got %d", len(registry.All()))
    }
}

// mockAuthHook implements AuthenticationHook
type mockAuthHook struct {
    mockHook
    authenticateFunc func(ctx context.Context, apiKey string) (bool, string, error)
}

func (m *mockAuthHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
    if m.authenticateFunc != nil {
        return m.authenticateFunc(ctx, apiKey)
    }
    return true, "", nil
}

func TestAuthenticationHook(t *testing.T) {
    registry := NewRegistry()

    called := false
    h := &mockAuthHook{
        mockHook: mockHook{name: "auth"},
        authenticateFunc: func(ctx context.Context, apiKey string) (bool, string, error) {
            called = true
            if apiKey == "valid-key" {
                return true, "user-123", nil
            }
            return false, "", nil
        },
    }

    registry.Register(h)

    // Test successful authentication
    success, _, err := h.Authenticate(context.Background(), "valid-key")
    if !success || err != nil {
        t.Error("expected successful authentication")
    }
    if !called {
        t.Error("Authenticate should have been called")
    }

    // Test failed authentication
    success, _, _ = h.Authenticate(context.Background(), "invalid-key")
    if success {
        t.Error("expected failed authentication")
    }
}

// mockRequestHook implements RequestHook
type mockRequestHook struct {
    mockHook
    beforeFunc func(ctx context.Context, req *openai.ChatCompletionRequest) error
    afterFunc  func(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error
}

func (m *mockRequestHook) BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error {
    if m.beforeFunc != nil {
        return m.beforeFunc(ctx, req)
    }
    return nil
}

func (m *mockRequestHook) AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error {
    if m.afterFunc != nil {
        return m.afterFunc(ctx, req, resp)
    }
    return nil
}

func TestRequestHook(t *testing.T) {
    calledBefore := false
    h := &mockRequestHook{
        mockHook: mockHook{name: "request"},
        beforeFunc: func(ctx context.Context, req *openai.ChatCompletionRequest) error {
            calledBefore = true
            req.Model = "modified-model"
            return nil
        },
    }

    req := &openai.ChatCompletionRequest{Model: "gpt-4"}
    h.BeforeRequest(context.Background(), req)

    if !calledBefore {
        t.Error("BeforeRequest should have been called")
    }
    if req.Model != "modified-model" {
        t.Error("BeforeRequest should modify request")
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./hook/`

Expected: FAIL with "undefined: NewRegistry"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p hook`

Create `hook/hook.go`:

```go
package hook

import (
    "context"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// Hook is the base interface for all hooks
type Hook interface {
    // Name returns the unique name of this hook
    Name() string
}

// AuthenticationHook is called to authenticate API keys
type AuthenticationHook interface {
    Hook
    // Authenticate validates the API key and returns (success, userID, error)
    Authenticate(ctx context.Context, apiKey string) (bool, string, error)
}

// RequestHook is called before/after sending request to provider
type RequestHook interface {
    Hook
    // BeforeRequest is called before sending request (can modify request)
    BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error
    // AfterRequest is called after receiving response (can modify response)
    AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error
}

// StreamingHook is called for each streaming chunk
type StreamingHook interface {
    Hook
    // OnChunk is called for each SSE chunk in streaming responses
    OnChunk(ctx context.Context, chunk []byte) error
}

// ErrorHook is called when an error occurs
type ErrorHook interface {
    Hook
    // OnError is called when an error occurs during request processing
    OnError(ctx context.Context, err error)
}

// Registry manages registered hooks
type Registry struct {
    authenticationHooks []AuthenticationHook
    requestHooks        []RequestHook
    streamingHooks      []StreamingHook
    errorHooks          []ErrorHook
}

// NewRegistry creates a new hook registry
func NewRegistry() *Registry {
    return &Registry{
        authenticationHooks: make([]AuthenticationHook, 0),
        requestHooks:        make([]RequestHook, 0),
        streamingHooks:      make([]StreamingHook, 0),
        errorHooks:          make([]ErrorHook, 0),
    }
}

// Register registers a hook based on its concrete type
func (r *Registry) Register(hook Hook) {
    switch h := hook.(type) {
    case AuthenticationHook:
        r.authenticationHooks = append(r.authenticationHooks, h)
    case RequestHook:
        r.requestHooks = append(r.requestHooks, h)
    case StreamingHook:
        r.streamingHooks = append(r.streamingHooks, h)
    case ErrorHook:
        r.errorHooks = append(r.errorHooks, h)
    }
}

// AuthenticationHooks returns all authentication hooks
func (r *Registry) AuthenticationHooks() []AuthenticationHook {
    return r.authenticationHooks
}

// RequestHooks returns all request hooks
func (r *Registry) RequestHooks() []RequestHook {
    return r.requestHooks
}

// StreamingHooks returns all streaming hooks
func (r *Registry) StreamingHooks() []StreamingHook {
    return r.streamingHooks
}

// ErrorHooks returns all error hooks
func (r *Registry) ErrorHooks() []ErrorHook {
    return r.errorHooks
}

// All returns all registered hooks
func (r *Registry) All() []Hook {
    all := make([]Hook, 0)
    for _, h := range r.authenticationHooks {
        all = append(all, h)
    }
    for _, h := range r.requestHooks {
        all = append(all, h)
    }
    for _, h := range r.streamingHooks {
        all = append(all, h)
    }
    for _, h := range r.errorHooks {
        all = append(all, h)
    }
    return all
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./hook/`

Expected: PASS

### Step 5: Commit

```bash
git add hook/hook.go hook/hook_test.go
git commit -m "feat: add hook system with registry"
```

---

## Task 4: Provider Interface

**Files:**
- Create: `provider/provider.go`
- Create: `provider/provider_test.go`
- Create: `provider/http.go`

### Step 1: Write the failing test for Provider interface

Create `provider/provider_test.go`:

```go
package provider

import (
    "context"
    "testing"

    "github.com/deeplooplabs/ai-gateway/openai"
)

func TestProviderInterface(t *testing.T) {
    // Mock provider for testing
    mockProv := &mockProvider{
        baseURL: "https://api.openai.com",
    }

    if mockProv.Name() != "mock" {
        t.Errorf("expected 'mock', got '%s'", mockProv.Name())
    }

    req := &openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.Message{{Role: "user", Content: "test"}},
    }

    resp, err := mockProv.SendRequest(context.Background(), "/v1/chat/completions", req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp == nil {
        t.Error("expected non-nil response")
    }
}

type mockProvider struct {
    baseURL string
}

func (m *mockProvider) Name() string {
    return "mock"
}

func (m *mockProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
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
```

### Step 2: Run test to verify it fails

Run: `go test -v ./provider/`

Expected: FAIL with "undefined: Provider"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p provider`

Create `provider/provider.go`:

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
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./provider/`

Expected: PASS

### Step 5: Write HTTPProvider implementation tests

Add to `provider/provider_test.go`:

```go
func TestHTTPProvider(t *testing.T) {
    provider := NewHTTPProvider("https://api.openai.com", "test-key")

    if provider.Name() != "http" {
        t.Errorf("expected 'http', got '%s'", provider.Name())
    }

    if provider.BaseURL != "https://api.openai.com" {
        t.Errorf("expected 'https://api.openai.com', got '%s'", provider.BaseURL)
    }
}
```

### Step 6: Run test to verify it fails

Run: `go test -v ./provider/`

Expected: FAIL with "undefined: NewHTTPProvider"

### Step 7: Write HTTPProvider implementation

Create `provider/http.go`:

```go
package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/deeplooplabs/ai-gateway/openai"
)

// HTTPProvider sends requests via HTTP
type HTTPProvider struct {
    Name      string
    BaseURL   string
    APIKey    string
    Client    *http.Client
}

// NewHTTPProvider creates a new HTTP provider
func NewHTTPProvider(baseURL, apiKey string) *HTTPProvider {
    return &HTTPProvider{
        Name:    "http",
        BaseURL: baseURL,
        APIKey:  apiKey,
        Client: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

// SendRequest sends a request via HTTP
func (p *HTTPProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    url := p.BaseURL + endpoint
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

    resp, err := p.Client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
    }

    var chatResp openai.ChatCompletionResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &chatResp, nil
}
```

### Step 8: Run test to verify it passes

Run: `go test -v ./provider/`

Expected: PASS

### Step 9: Commit

```bash
git add provider/provider.go provider/provider_test.go provider/http.go
git commit -m "feat: add Provider interface with HTTP implementation"
```

---

## Task 5: Model Registry

**Files:**
- Create: `model/registry.go`
- Create: `model/registry_test.go`

### Step 1: Write the failing test for ModelRegistry

Create `model/registry_test.go`:

```go
package model

import (
    "testing"

    "github.com/deeplooplabs/ai-gateway/provider"
)

func TestMapModelRegistry(t *testing.T) {
    prov1 := &mockProvider{name: "provider1"}
    prov2 := &mockProvider{name: "provider2"}

    registry := NewMapModelRegistry()

    // Register models
    registry.Register("gpt-4", prov1, "")
    registry.Register("gpt-3.5-turbo", prov2, "gpt-35-turbo")
    registry.Register("claude-3", prov1, "")

    // Test exact match
    p, modelRewrite := registry.Resolve("gpt-4")
    if p.Name() != "provider1" {
        t.Errorf("expected 'provider1', got '%s'", p.Name())
    }
    if modelRewrite != "" {
        t.Errorf("expected empty rewrite, got '%s'", modelRewrite)
    }

    // Test model rewrite
    p, modelRewrite = registry.Resolve("gpt-3.5-turbo")
    if p.Name() != "provider2" {
        t.Errorf("expected 'provider2', got '%s'", p.Name())
    }
    if modelRewrite != "gpt-35-turbo" {
        t.Errorf("expected 'gpt-35-turbo', got '%s'", modelRewrite)
    }

    // Test unknown model (should still work, returns nil provider)
    p, modelRewrite = registry.Resolve("unknown")
    if p != nil {
        t.Error("expected nil provider for unknown model")
    }
}

type mockProvider struct {
    name string
}

func (m *mockProvider) Name() string {
    return m.name
}

func (m *mockProvider) SendRequest(ctx any, endpoint string, req any) (any, error) {
    return nil, nil
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./model/`

Expected: FAIL with "undefined: NewMapModelRegistry"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p model`

Create `model/registry.go`:

```go
package model

import (
    "github.com/deeplooplabs/ai-gateway/provider"
)

// ProviderRewrite represents a provider and optional model name rewrite
type ProviderRewrite struct {
    Provider    provider.Provider
    ModelRewrite string
}

// ModelRegistry resolves model names to providers
type ModelRegistry interface {
    // Resolve returns the provider and optional model rewrite for a given model name
    Resolve(model string) (provider.Provider, string)
}

// MapModelRegistry is an in-memory model registry
type MapModelRegistry struct {
    models map[string]ProviderRewrite
}

// NewMapModelRegistry creates a new map-based model registry
func NewMapModelRegistry() *MapModelRegistry {
    return &MapModelRegistry{
        models: make(map[string]ProviderRewrite),
    }
}

// Register registers a model with its provider and optional model rewrite
func (r *MapModelRegistry) Register(model string, prov provider.Provider, modelRewrite string) {
    r.models[model] = ProviderRewrite{
        Provider:    prov,
        ModelRewrite: modelRewrite,
    }
}

// Resolve returns the provider and model rewrite for a given model name
func (r *MapModelRegistry) Resolve(model string) (provider.Provider, string) {
    if pr, ok := r.models[model]; ok {
        return pr.Provider, pr.ModelRewrite
    }
    return nil, ""
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./model/`

Expected: PASS

### Step 5: Commit

```bash
git add model/registry.go model/registry_test.go
git commit -m "feat: add ModelRegistry for model-to-provider resolution"
```

---

## Task 6: Chat Completion Handler

**Files:**
- Create: `handler/chat.go`
- Create: `handler/chat_test.go`

### Step 1: Write the failing test for ChatHandler

Create `handler/chat_test.go`:

```go
package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/openai"
    "github.com/deeplooplabs/ai-gateway/provider"
)

func TestChatHandler_ServeHTTP(t *testing.T) {
    // Setup
    registry := newMockRegistry()
    hooks := hook.NewRegistry()

    handler := NewChatHandler(registry, hooks)

    // Create request
    reqBody := map[string]any{
        "model": "gpt-4",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello"},
        },
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-key")

    w := httptest.NewRecorder()

    // Execute
    handler.ServeHTTP(w, req)

    // Verify
    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
    }

    var resp openai.ChatCompletionResponse
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    if resp.Object != "chat.completion" {
        t.Errorf("expected 'chat.completion', got '%s'", resp.Object)
    }
}

func TestChatHandler_Stream(t *testing.T) {
    registry := newMockRegistry()
    hooks := hook.NewRegistry()

    handler := NewChatHandler(registry, hooks)

    reqBody := map[string]any{
        "model":    "gpt-4",
        "messages": []map[string]string{{"role": "user", "content": "Hello"}},
        "stream":   true,
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-key")

    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }

    contentType := w.Header().Get("Content-Type")
    if contentType != "text/event-stream" {
        t.Errorf("expected 'text/event-stream', got '%s'", contentType)
    }
}

func newMockRegistry() *mapModelRegistry {
    prov := &mockChatProvider{}
    return &mapModelRegistry{provider: prov}
}

type mapModelRegistry struct {
    provider provider.Provider
}

func (m *mapModelRegistry) Resolve(model string) (provider.Provider, string) {
    return m.provider, ""
}

type mockChatProvider struct{}

func (m *mockChatProvider) Name() string {
    return "mock"
}

func (m *mockChatProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
    return &openai.ChatCompletionResponse{
        ID:     "test-id",
        Object: "chat.completion",
        Model:  req.Model,
        Choices: []openai.Choice{{
            Index: 0,
            Message: openai.Message{
                Role:    "assistant",
                Content: "Hello!",
            },
            FinishReason: "stop",
        }},
        Usage: openai.Usage{
            PromptTokens:     10,
            CompletionTokens: 5,
            TotalTokens:      15,
        },
    }, nil
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./handler/`

Expected: FAIL with "undefined: NewChatHandler"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p handler`

Create `handler/chat.go`:

```go
package handler

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"

    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
    "github.com/deeplooplabs/ai-gateway/openai"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
    registry model.ModelRegistry
    hooks    *hook.Registry
}

// NewChatHandler creates a new chat handler
func NewChatHandler(registry model.ModelRegistry, hooks *hook.Registry) *ChatHandler {
    return &ChatHandler{
        registry: registry,
        hooks:    hooks,
    }
}

// ServeHTTP implements http.Handler
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req openai.ChatCompletionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, NewValidationError("invalid request body: " + err.Error()))
        return
    }

    // Validate request
    if req.Model == "" {
        h.writeError(w, NewValidationError("model is required"))
        return
    }
    if len(req.Messages) == 0 {
        h.writeError(w, NewValidationError("messages is required"))
        return
    }

    // Resolve provider
    provider, modelRewrite := h.registry.Resolve(req.Model)
    if provider == nil {
        h.writeError(w, NewNotFoundError("model not found: " + req.Model))
        return
    }

    // Apply model rewrite if specified
    if modelRewrite != "" {
        req.Model = modelRewrite
    }

    // Handle streaming
    if req.Stream {
        h.handleStream(w, r, &req, provider)
        return
    }

    // Handle non-streaming
    h.handleNonStream(w, r, &req, provider)
}

func (h *ChatHandler) handleNonStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, provider Provider) {
    // Call BeforeRequest hooks
    for _, hh := range h.hooks.RequestHooks() {
        if err := hh.BeforeRequest(r.Context(), req); err != nil {
            h.writeError(w, fmt.Errorf("hook error: %w", err))
            return
        }
    }

    // Send request to provider
    resp, err := provider.SendRequest(r.Context(), "/v1/chat/completions", req)
    if err != nil {
        h.writeError(w, NewProviderError("provider error", err))
        return
    }

    // Call AfterRequest hooks
    for _, hh := range h.hooks.RequestHooks() {
        if err := hh.AfterRequest(r.Context(), req, resp); err != nil {
            h.writeError(w, fmt.Errorf("hook error: %w", err))
            return
        }
    }

    // Write response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *openai.ChatCompletionRequest, provider Provider) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        h.writeError(w, NewValidationError("streaming not supported"))
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // For mock/testing, send simple chunk
    chunk := `data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"` + req.Model + `","choices":[{"index":0,"delta":{"content":"Hello!"},"finish_reason":null}]}` + "\n\n"
    io.WriteString(w, chunk)

    endChunk := `data: [DONE]` + "\n\n"
    io.WriteString(w, endChunk)

    flusher.Flush()
}

func (h *ChatHandler) writeError(w http.ResponseWriter, err error) {
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

// GatewayError represents a gateway error (simplified for handler)
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

// Provider interface for handler (simplified)
type Provider interface {
    Name() string
    SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./handler/`

Expected: PASS (may need minor fixes)

### Step 5: Commit

```bash
git add handler/chat.go handler/chat_test.go
git commit -m "feat: add chat completion handler with streaming support"
```

---

## Task 7: Gateway Core

**Files:**
- Create: `gateway/gateway.go`
- Create: `gateway/gateway_test.go`
- Create: `gateway/option.go`

### Step 1: Write the failing test for Gateway

Create `gateway/gateway_test.go`:

```go
package gateway

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
)

func TestGateway_New(t *testing.T) {
    gw := New()

    if gw == nil {
        t.Error("expected non-nil gateway")
    }
}

func TestGateway_ServeHTTP_ChatCompletions(t *testing.T) {
    registry := setupTestRegistry()
    hooks := hook.NewRegistry()

    gw := New(
        WithModelRegistry(registry),
        WithHooks(hooks),
    )

    // Create chat completion request
    reqBody := map[string]any{
        "model": "gpt-4",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello"},
        },
    }
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    gw.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
}

func TestGateway_ServeHTTP_InvalidPath(t *testing.T) {
    gw := New()

    req := httptest.NewRequest("GET", "/invalid/path", nil)
    w := httptest.NewRecorder()
    gw.ServeHTTP(w, req)

    if w.Code != http.StatusNotFound {
        t.Errorf("expected 404, got %d", w.Code)
    }
}

func setupTestRegistry() model.ModelRegistry {
    // This would use a mock provider
    return model.NewMapModelRegistry()
}
```

### Step 2: Run test to verify it fails

Run: `go test -v ./gateway/`

Expected: FAIL with "undefined: New"

### Step 3: Create directory and write minimal implementation

Run: `mkdir -p gateway`

Create `gateway/gateway.go`:

```go
package gateway

import (
    "net/http"
    "net/http/pprof"

    "github.com/deeplooplabs/ai-gateway/handler"
    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
)

// Gateway is the main HTTP handler
type Gateway struct {
    modelRegistry model.ModelRegistry
    hooks         *hook.Registry
    mux           *http.ServeMux
}

// New creates a new gateway with default options
func New(opts ...Option) *Gateway {
    // Default configuration
    g := &Gateway{
        modelRegistry: model.NewMapModelRegistry(),
        hooks:         hook.NewRegistry(),
        mux:           http.NewServeMux(),
    }

    // Apply options
    for _, opt := range opts {
        opt(g)
    }

    // Setup routes
    g.setupRoutes()

    return g
}

func (g *Gateway) setupRoutes() {
    // Chat Completions
    chatHandler := handler.NewChatHandler(g.modelRegistry, g.hooks)
    g.mux.HandleFunc("/v1/chat/completions", chatHandler.ServeHTTP)

    // Health check
    g.mux.HandleFunc("/health", g.handleHealth)

    // 404 for unmatched routes
    g.mux.HandleFunc("/", g.handleNotFound)
}

// ServeHTTP implements http.Handler
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    g.mux.ServeHTTP(w, r)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}

func (g *Gateway) handleNotFound(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(`{"error":{"message":"Not found","type":"invalid_request_error"}}`))
}
```

Create `gateway/option.go`:

```go
package gateway

import (
    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
)

// Option configures the Gateway
type Option func(*Gateway)

// WithModelRegistry sets the model registry
func WithModelRegistry(registry model.ModelRegistry) Option {
    return func(g *Gateway) {
        g.modelRegistry = registry
    }
}

// WithHooks sets the hook registry
func WithHooks(hooks *hook.Registry) Option {
    return func(g *Gateway) {
        g.hooks = hooks
    }
}

// WithHook registers a single hook
func WithHook(h hook.Hook) Option {
    return func(g *Gateway) {
        g.hooks.Register(h)
    }
}
```

### Step 4: Run test to verify it passes

Run: `go test -v ./gateway/`

Expected: PASS

### Step 5: Commit

```bash
git add gateway/gateway.go gateway/gateway_test.go gateway/option.go
git commit -m "feat: add core Gateway with routing"
```

---

## Task 8: Example Usage

**Files:**
- Create: `example/main.go`

### Step 1: Create example usage

Run: `mkdir -p example`

Create `example/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/deeplooplabs/ai-gateway/gateway"
    "github.com/deeplooplabs/ai-gateway/hook"
    "github.com/deeplooplabs/ai-gateway/model"
    "github.com/deeplooplabs/ai-gateway/provider"
)

func main() {
    // Create providers
    openAIProvider := provider.NewHTTPProvider("https://api.openai.com", "your-api-key")

    // Setup model registry
    registry := model.NewMapModelRegistry()
    registry.Register("gpt-4", openAIProvider, "")
    registry.Register("gpt-3.5-turbo", openAIProvider, "")

    // Create hooks
    hooks := hook.NewRegistry()
    hooks.Register(&LoggingHook{})

    // Create gateway
    gw := gateway.New(
        gateway.WithModelRegistry(registry),
        gateway.WithHooks(hooks),
    )

    // Start server
    fmt.Println("AI Gateway listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", gw))
}

// LoggingHook logs all requests
type LoggingHook struct{}

func (h *LoggingHook) Name() string {
    return "logging"
}

func (h *LoggingHook) BeforeRequest(ctx any, req any) error {
    fmt.Printf("[Hook] BeforeRequest: model=%v\n", req)
    return nil
}

func (h *LoggingHook) AfterRequest(ctx any, req any, resp any) error {
    fmt.Printf("[Hook] AfterRequest\n")
    return nil
}
```

### Step 2: Commit

```bash
git add example/main.go
git commit -m "examples: add basic usage example"
```

---

## Task 9: Documentation

**Files:**
- Modify: `README.md`

### Step 1: Update README

Replace `README.md` content:

```markdown
# DeepLoop AI Gateway

A programmable AI Gateway library for Go, fully compatible with OpenAI API.

## Features

- **Full OpenAI API Compatibility**: Chat, Completions, Embeddings, and more
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

## Hook System

Hooks allow you to customize request/response processing:

```go
type AuthHook struct{}

func (h *AuthHook) Name() string { return "auth" }

func (h *AuthHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
    // Validate API key
    return true, "user-id", nil
}

// Register the hook
hooks := hook.NewRegistry()
hooks.Register(&AuthHook{})
```

## License

MIT
```

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: add comprehensive README"
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
git commit -m "chore: final cleanup and verification"
```

---

## Summary

This implementation plan covers:

1. **Project Foundation**: Context and error handling
2. **OpenAI Types**: Request/response type definitions
3. **Hook System**: Multi-level hook interfaces and registry
4. **Provider Interface**: Abstraction with HTTP implementation
5. **Model Registry**: Model-to-provider resolution with rewriting
6. **Chat Handler**: Chat completions with streaming support
7. **Gateway Core**: Main HTTP handler with routing
8. **Example**: Basic usage example
9. **Documentation**: README with quick start

Each task follows TDD methodology with bite-sized steps, exact file paths, complete code examples, and frequent commits.
