# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**DeepLoop AI Gateway** is a programmable AI Gateway library written in Go that provides **OpenAI API compatibility** and implements the [OpenResponses specification](https://www.openresponses.org/). It is designed as an **embeddable library** (not a standalone service) that can be integrated into Go applications to provide unified access to multiple LLM providers.

### Key Design Principle

The gateway supports **two API specifications simultaneously**:
- **OpenAI API** (`/v1/chat/completions`, `/v1/embeddings`, `/v1/images/generations`)
- **OpenResponses API** (`/v1/responses` with semantic streaming events)

## Development Commands

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./handler/...
go test ./openresponses/...
go test ./provider/...

# Run tests with verbose output
go test -v ./...

# Build the library
go build .

# Run the example server
go run example/main.go

# Install dependencies
go mod tidy

# Run specific test
go test -run TestChatHandler -v ./handler/
```

## Architecture

### Layer Structure

```
┌─────────────────────────────────────────┐
│           HTTP Layer                    │
│  (gateway/gateway.go - http.Handler)   │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│         Handler Layer                   │
│  (handler/*.go)                         │
│  - ChatHandler (OpenAI)                 │
│  - ResponsesHandler (OpenResponses)     │
│  - EmbeddingsHandler                    │
│  - ImagesHandler                        │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│         Hook System                     │
│  (hook/hook.go)                         │
│  - AuthenticationHook                   │
│  - RequestHook (before/after)           │
│  - StreamingHook (per-chunk)            │
│  - ErrorHook                            │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Model Registry                     │
│  (model/registry.go)                    │
│  - Maps model names to providers        │
│  - Supports model name rewriting        │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Provider Layer                     │
│  (provider/provider.go)                 │
│  - HTTPProvider (generic REST)          │
│  - GeminiHTTPProvider                   │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Conversion Layer                   │
│  - openresponses/converter.go           │
│  - provider/converter.go                │
│  - Bidirectional OpenAI ↔ OR conversion │
└─────────────────────────────────────────┘
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `gateway/` | Main HTTP handler implementing `http.Handler`. Routes requests to appropriate handlers. |
| `handler/` | HTTP handlers for each API endpoint (chat, responses, embeddings, images). |
| `openresponses/` | **OpenResponses types** and streaming event implementation. |
| `provider/` | Abstract interface for LLM providers with unified Request/Response types. |
| `provider/openai/` | **OpenAI types** - canonical location for OpenAI API schemas. |
| `hook/` | Extensible hook system with 4 hook types. |
| `model/` | Model registry that maps model names to providers. |
| `cache/` | LRU cache for response caching with TTL support. |
| `ratelimit/` | Token bucket rate limiter for request throttling. |
| `quota/` | Token usage quota tracking and enforcement. |
| `loadbalancer/` | Multi-provider load balancing with health checks. |
| `e2e/` | End-to-end tests using OpenAI client library. |

## OpenResponses Implementation

### Endpoint

The gateway implements `POST /v1/responses` as specified in [OpenResponses](https://www.openresponses.org/).

### Request/Response Types

All OpenResponses types are defined in `openresponses/types.go`:

```go
// Request
type CreateRequest struct {
    Model              string
    Input              InputParam  // string or []MessageItemParam
    PreviousResponseID string
    Tools              []Tool
    ToolChoice         ToolChoiceParam
    Stream             *bool
    Temperature        *float64
    MaxOutputTokens    *int
    Truncation         TruncationEnum
    // ...
}

// Response
type Response struct {
    ID          string
    Object      string  // "response"
    Status      ResponseStatusEnum  // in_progress, completed, failed, incomplete
    CreatedAt   int64
    CompletedAt *int64
    Model       string
    Output      []ItemField  // MessageItem, FunctionCallItem, etc.
    Usage       *Usage
    Error       *Error
    // ...
}
```

### Streaming Events

OpenResponses streaming uses **semantic events**, not raw deltas. All streaming events are defined in `openresponses/streaming.go`:

**State Machine Events:**
```go
NewResponseCreatedEvent(sequence, responseID)
NewResponseInProgressEvent(sequence)
NewResponseCompletedEvent(sequence, response)
NewResponseFailedEvent(sequence, errorType, message)
NewResponseIncompleteEvent(sequence, details)
```

**Item Events:**
```go
NewResponseOutputItemAddedEvent(sequence, index, item)
NewResponseOutputItemDoneEvent(sequence, index, item)
```

**Content Events:**
```go
NewResponseOutputTextDeltaEvent(sequence, itemID, outputIndex, contentIndex, delta)
NewResponseOutputTextDoneEvent(sequence, itemID, outputIndex, contentIndex, text)
NewResponseContentPartAddedEvent(sequence, itemID, outputIndex, contentIndex, content)
```

### Streaming Protocol

- **Content-Type**: `text/event-stream`
- **Termination**: Literal `[DONE]` marker
- **Event field**: Matches `type` in event body
- **Sequence number**: Required for ordering

Example streaming events:
```
event: response.created
data: {"type":"response.created","response_id":"resp_abc123",...}

event: response.in_progress
data: {"type":"response.in_progress","sequence_number":1}

event: response.output_item.added
data: {"type":"response.output_item.added","sequence_number":2,...}

event: response.output_text.delta
data: {"type":"response.output_text.delta","sequence_number":3,"delta":"Hello"}

event: response.completed
data: {"type":"response.completed","sequence_number":10,...}

data: [DONE]
```

### Error Format

OpenResponses errors follow this format:

```json
{
  "error": {
    "type": "invalid_request_error|server_error|not_found|model_error|too_many_requests",
    "message": "Human-readable description",
    "param": "optional_parameter_name",
    "code": "optional_specific_code"
  }
}
```

Error implementations in `error.go`:
- `NewValidationError()` → `invalid_request_error` (400)
- `NewNotFoundError()` → `not_found` (404)
- `NewAuthenticationError()` → `authentication_error` (401)
- `NewRateLimitError()` → `rate_limit_error` (429)
- `NewServerError()` → `server_error` (500)

## OpenAI Compatibility

### OpenAI Types

All OpenAI request/response types are defined in `provider/openai/types.go`. This is the canonical location for OpenAI API schemas.

### Supported Endpoints

| Endpoint | Handler | Status |
|----------|---------|--------|
| `/v1/chat/completions` | `ChatHandler` | ✅ Full support |
| `/v1/embeddings` | `EmbeddingsHandler` | ✅ Full support |
| `/v1/images/generations` | `ImagesHandler` | ✅ Full support |
| `/v1/responses` | `ResponsesHandler` | ✅ Full support (OpenResponses) |
| `/v1/models` | `ModelsHandler` | ✅ List available models |
| `/health` | Built-in | ✅ Health check endpoint |
| `/metrics` | Prometheus | ✅ Metrics (if enabled) |

## Conversion Between Formats

The gateway supports bidirectional conversion between OpenAI and OpenResponses formats:

**Implemented in `openresponses/converter.go`:**
- `RequestToChatCompletion()` - OR Request → OpenAI Request
- `ChatCompletionToResponse()` - OpenAI Response → OR Response
- `ResponseToChatCompletion()` - OR Response → OpenAI Response (reverse)
- `StreamingChunkToEvents()` - OpenAI chunks → OR semantic events

**Implemented in `provider/response.go`:**
- `GetChatCompletion()` - Returns OpenAI format (converts from OR if needed)
- `GetORResponse()` - Returns OR format (converts from OpenAI if needed)

## Hook System

Hooks are registered via `hook.NewRegistry()` and passed to gateway options:

```go
hooks := hook.NewRegistry()
hooks.Register(&MyAuthHook{}, &MyLoggingHook{})
gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithHooks(hooks),
)
```

**Important**: Hook execution order is the order they were registered. Context values (like `tenant_id` from authentication) are available to downstream hooks via `context.Context`.

### Hook Types

| Hook Type | Interface | Called When |
|-----------|-----------|-------------|
| `AuthenticationHook` | `Authenticate(ctx, apiKey) (success, tenantID, err)` | Before request processing |
| `RequestHook` | `BeforeRequest(ctx, req)`, `AfterRequest(ctx, req, resp)` | Before/after provider call |
| `StreamingHook` | `OnChunk(ctx, chunk) (modifiedChunk, err)` | For each streaming chunk |
| `ErrorHook` | `OnError(ctx, err)` | On any error |

## Provider Configuration

### HTTP Provider

The `HTTPProvider` is a generic provider for OpenAI-compatible REST APIs:

```go
// Basic configuration
provider := provider.NewHTTPProviderWithBaseURL(
    "https://api.openai.com/v1",
    "your-api-key",
)

// With BasePath (for APIs that include /v1 in base URL)
provider := provider.NewHTTPProviderWithBaseURLAndPath(
    "https://api.siliconflow.cn/v1",  // BaseURL includes /v1
    "your-api-key",
    "/v1",  // Strip /v1 from endpoint
)

// Full configuration
config := provider.NewProviderConfig("my-provider").
    WithBaseURL("https://api.example.com/v1").
    WithBasePath("/v1").  // Strip from endpoint before appending
    WithAPIKey("your-key").
    WithAPIType(provider.APITypeAll).  // Support both ChatCompletions and Responses
    WithTimeout(30 * time.Second)

provider := provider.NewHTTPProvider(config)
```

### Provider Interface

All providers implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    SupportedAPIs() APIType
    SendRequest(ctx context.Context, req *Request) (*Response, error)
}
```

## Model Registry

The model registry maps model names to providers with optional transformations:

```go
registry := model.NewMapModelRegistry()

// Simple registration
registry.Register("gpt-4", provider)

// With model name rewrite (for provider-specific model names)
registry.RegisterWithOptions("gpt-4", provider,
    model.WithModelRewrite("deepseek-ai/DeepSeek-V3"),
    model.WithPreferredAPI(provider.APITypeChatCompletions),
)

// Resolve returns (provider, rewrittenModelName)
prov, modelName := registry.Resolve("gpt-4")
```

## Streaming Implementation

### OpenAI Streaming (SSE with raw deltas)

```go
// Provider returns chunks via channel
for chunk := range resp.Chunks {
    if chunk.Type == provider.ChunkTypeOpenAI {
        // Raw OpenAI SSE chunk
        data := chunk.OpenAI.Data
    }
}
```

### OpenResponses Streaming (semantic events)

```go
// Converter transforms OpenAI chunks to OR events
events := converter.StreamingChunkToEvents(chunk.Data, &seq, itemID, outputIndex)
for _, event := range events {
    writer.WriteEvent(event)  // Writes proper SSE format
}
```

## Adding a New Provider

1. Implement the `provider.Provider` interface
2. Optionally implement streaming by returning channels for `Chunk` and errors
3. Register models with the provider in the model registry
4. Use `provider.NewHTTPProvider()` for standard OpenAI-compatible APIs

## Testing Strategy

- Unit tests for each package alongside source files (`*_test.go`)
- Handler tests mock provider responses
- Provider tests use recorded responses or test servers
- Hook tests verify execution order and context propagation
- Conversion tests verify bidirectional OpenAI ↔ OR conversion

## Configuration

The gateway uses functional options for configuration:

```go
gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithHooks(hooks),
)
```

All available options are in `gateway/option.go`.

### Available Gateway Options

| Option | Description |
|--------|-------------|
| `WithModelRegistry(registry)` | Set the model registry |
| `WithHooks(hooks)` | Set the hook registry |
| `WithHook(hook)` | Register a single hook |
| `WithCORS(config)` | Enable CORS with configuration |
| `WithMetrics(namespace)` | Enable Prometheus metrics |
| `WithCache(cache)` | Enable response caching |
| `WithRateLimiter(limiter)` | Enable rate limiting |

## Advanced Features

### Response Caching

Cache LLM responses to reduce latency and costs:

```go
import "github.com/deeplooplabs/ai-gateway/cache"

// Create LRU cache
cacheImpl := cache.NewLRUCache(&cache.Config{
    MaxSize:    100 * 1024 * 1024, // 100MB
    MaxItems:   10000,
    DefaultTTL: 5 * time.Minute,
    Enabled:    true,
})

gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithCache(cacheImpl),
)
```

**Cache Interface:**
- `Get(ctx, key)` - Retrieve cached value
- `Set(ctx, key, value, ttl)` - Store value with TTL
- `Delete(ctx, key)` - Remove value
- `Clear(ctx)` - Clear all values
- `Stats()` - Get cache statistics (hits, misses, size, items)

### Rate Limiting

Throttle requests using token bucket algorithm:

```go
import "github.com/deeplooplabs/ai-gateway/ratelimit"

// Create rate limiter
limiter := ratelimit.NewTokenBucket(&ratelimit.Config{
    RequestsPerSecond: 100, // 100 RPS
    Burst:             200, // Allow bursts up to 200
    Enabled:           true,
})

gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithRateLimiter(limiter),
)
```

**Rate Limiter Interface:**
- `Allow(ctx, key)` - Check if single request is allowed
- `AllowN(ctx, key, n)` - Check if N requests are allowed
- `Reset(ctx, key)` - Reset limiter for key

### Quota Management

Track and enforce token usage quotas per tenant:

```go
import "github.com/deeplooplabs/ai-gateway/quota"

// Create quota manager
quotaMgr := quota.NewManager(&quota.Config{
    DefaultQuota: 1000000, // 1M tokens per tenant
    ResetPeriod:  quota.Monthly,
})

// Set specific quota for a tenant
quotaMgr.SetQuota(ctx, "tenant-123", 5000000)

// Check quota before request
allowed, usage, err := quotaMgr.CheckQuota(ctx, tenantID)

// Record usage after request
quotaMgr.RecordUsage(ctx, tenantID, inputTokens, outputTokens, totalTokens)
```

**Quota Manager Interface:**
- `RecordUsage(ctx, tenantID, inputTokens, outputTokens, totalTokens)` - Record usage
- `CheckQuota(ctx, tenantID)` - Check if tenant has remaining quota
- `GetUsage(ctx, tenantID)` - Get current usage
- `SetQuota(ctx, tenantID, limit)` - Set quota limit
- `ResetUsage(ctx, tenantID)` - Reset usage for tenant
- `ResetAll(ctx)` - Reset all tenant usage

**Reset Periods:** `Hourly`, `Daily`, `Weekly`, `Monthly`, `Never`

### Load Balancing

Distribute requests across multiple providers:

```go
import "github.com/deeplooplabs/ai-gateway/loadbalancer"

// Create load-balanced provider
lb := loadbalancer.New("my-balanced-provider", loadbalancer.RoundRobin)
lb.AddProvider(provider1, 1) // weight: 1
lb.AddProvider(provider2, 2) // weight: 2 (gets 2x traffic)

// Enable health checks
lb.EnableHealthChecks(30 * time.Second)

// Register in model registry
registry.Register("gpt-4", lb)
```

**Load Balancing Strategies:**
- `RoundRobin` - Evenly distribute across providers
- `Random` - Random provider selection
- `WeightedRandom` - Random selection weighted by provider weight
- `LeastConnections` - Select provider with fewest active requests

**Health Checks:**
- Automatic health monitoring at configurable intervals
- Unhealthy providers are automatically removed from rotation
- Providers are re-added when they become healthy again

### Metrics

Expose Prometheus metrics for monitoring:

```go
gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithMetrics("ai_gateway"), // namespace prefix
)

// Metrics available at /metrics endpoint
```

**Exported Metrics:**
- Request counts by endpoint, model, and status
- Request duration histograms
- Token usage by model
- Cache hit/miss rates
- Rate limiter rejections
- Provider health status

### CORS

Enable Cross-Origin Resource Sharing:

```go
import "github.com/deeplooplabs/ai-gateway/gateway"

cors := gateway.DefaultCORSConfig() // Allows all origins
// Or customize:
cors := &gateway.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           time.Hour,
}

gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithCORS(cors),
)
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run E2E tests (uses OpenAI client)
go test ./e2e/...

# Skip E2E tests in short mode
go test -short ./...
```

### E2E Test Structure

E2E tests in `e2e/` use the real OpenAI client library (`github.com/sashabaranov/go-openai`) to test the gateway as a black box:

```go
// Setup test environment with mock provider
env := NewTestEnvironment(t)

// Configure mock response
env.MockProvider.SetChatResponse(&openai.ChatCompletionResponse{...})

// Use real OpenAI client pointing to test server
client := env.Client
resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model: "gpt-4",
    Messages: []openai.ChatCompletionMessage{...},
})
```

**Test Categories:**
- Chat completions (streaming and non-streaming)
- OpenResponses API (streaming and non-streaming)
- Embeddings
- Images
- Models endpoint
- Error handling
