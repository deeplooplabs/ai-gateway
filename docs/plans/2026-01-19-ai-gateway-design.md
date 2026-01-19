# AI Gateway Design

**Date**: 2026-01-19
**Status**: Draft

## Overview

A programmable AI Gateway library for Go, fully compatible with OpenAI API. Designed to be embedded into third-party services as a library, providing a flexible hook system for request/response processing.

## Core Principles

- **Library-first**: Can be embedded as a library, not just a standalone service
- **Full OpenAI compatibility**: Chat, Completions, Embeddings, Images, Audio, Files, etc.
- **Multi-level hooks**: Extensible at authentication, routing, request, response, chunk, and error stages
- **Provider abstraction**: Interface-based provider management, user-controlled implementation
- **Parse-reconstruct mode**: Parse requests, process through hooks, reconstruct and forward

## Architecture

### 1. Overall Architecture

The core is the `Gateway` struct implementing `http.Handler`, directly embeddable into Go HTTP services.

**Layers**:

- **Routing Layer**: Routes incoming requests to appropriate handlers based on OpenAI API path format `/v1/{resource}` and HTTP method (GET/POST/DELETE).

- **Handler Layer**: Each API endpoint has a corresponding handler responsible for request parsing and response assembly. Handlers are extensible; users can register custom handlers.

- **Provider Layer**: Defines `Provider` interface with `SendRequest(ctx, req) (*Response, error)`. Library provides default `HTTPProvider`; users can implement custom providers (internal RPC, mock, etc.).

- **Hook Layer**: Cross-cutting hook system providing extension points at key lifecycle stages.

### 2. Hook System

```go
type Hook interface {
    Name() string
}

type AuthenticationHook interface {
    Hook
    Authenticate(ctx *Context, apiKey string) (bool, string, error)
}

type RequestHook interface {
    Hook
    BeforeRequest(ctx *Context, req *openai.Request) error
    AfterRequest(ctx *Context, req *openai.Request, resp *openai.Response) error
}

type StreamingHook interface {
    Hook
    OnChunk(ctx *Context, chunk []byte) error
}

type ErrorHook interface {
    Hook
    OnError(ctx *Context, err error)
}
```

Hook execution is ordered by priority; same priority hooks execute in registration order. Hooks can return error to interrupt flow or special flag to skip subsequent hooks.

**Context** is passed throughout the request lifecycle:
```go
type Context struct {
    RequestID   string
    StartTime   time.Time
    OriginalReq *http.Request
    Metadata    map[string]any
    Provider    Provider
}
```

### 3. Request Flow

1. HTTP request enters → Generate RequestID → Create Context
2. Execute `AuthenticationHook` for API Key validation
3. Route to handler, parse request body to OpenAI structs
4. Execute `BeforeRequest` hooks (can modify request)
5. Provider sends request
6. Receive response:
   - Non-streaming: Parse fully, execute `AfterRequest` hooks
   - Streaming: Execute `OnChunk` hook per chunk, forward immediately via flusher
7. On error, execute `ErrorHook`
8. Record metrics, return response

### 4. Provider Interface & Model Management

```go
type Provider interface {
    Name() string
    SendRequest(ctx context.Context, endpoint string, req *Request) (*Response, error)
    SendRequestStream(ctx context.Context, endpoint string, req *Request) (<-chan StreamChunk, <-chan error)
}

type HTTPProvider struct {
    BaseURL    string
    HTTPClient *http.Client
    APIKey     string
}
```

**ModelRegistry**:
```go
type ModelRegistry interface {
    Resolve(model string) (Provider, string)
}
```

Returns: specific Provider instance, and optional model rewrite name.

Users can implement:
- Dynamic routing by cost/latency
- Model rewriting (e.g., `gpt-4` → `claude-3-opus`)
- Multi-tenancy (different API keys map to different providers)

Default `MapModelRegistry` provided.

### 5. Error Handling & Observability

**Error Handling**: Unified `GatewayError` type with error code, message, and original error. Hooks can modify errors (e.g., redact sensitive info). All errors converted to OpenAI-format JSON for compatibility.

**Observability**: Built-in OpenTelemetry support (opt-in). Each request creates a span recording:
- Request path, model name
- Provider call duration
- Token usage (parsed from response)
- Error information

Default `MetricsHook` outputs Prometheus metrics:
- `ai_gateway_requests_total`
- `ai_gateway_duration_seconds`
- `ai_gateway_tokens_total`

Users enable via `WithTracing()` and `WithMetrics()` options, or implement custom hooks for other systems.

### 6. Package Structure

```
ai-gateway/
├── gateway/          # Core Gateway struct
├── handler/          # API endpoint handlers
│   ├── chat.go
│   ├── embeddings.go
│   ├── images.go
│   └── ...
├── hook/             # Hook interfaces and implementations
│   ├── auth.go
│   ├── request.go
│   ├── streaming.go
│   └── telemetry.go
├── provider/         # Provider interface and implementations
├── model/            # ModelRegistry and logic
├── openai/           # OpenAI request/response types
├── middleware/       # HTTP middleware (recover, logger)
├── config/           # Gateway configuration options
└── context.go        # Context definition
```

### 7. Usage Example

```go
import "github.com/deeplooplabs/ai-gateway/gateway"

gw := gateway.New(
    gateway.WithModelRegistry(myRegistry),
    gateway.WithHooks(authHook, metricsHook),
)
http.Handle("/v1/", gw)
```

## Hook Use Cases

- **API Key validation and replacement**: Verify keys, map to upstream provider keys
- **Token usage statistics**: Track and report usage per user/model
- **Open Telemetry integration**: Custom traces and metrics
- **Provider/model management**: Dynamic routing, model rewriting
- **Request/response transformation**: Modify prompts, filter responses

## Next Steps

1. Implement core Gateway structure and routing
2. Define OpenAI request/response types
3. Implement Hook system interfaces
4. Implement HTTPProvider and ModelRegistry
5. Add handlers for each API endpoint
6. Add telemetry support
