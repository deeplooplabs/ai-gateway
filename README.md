# DeepLoop AI Gateway

A programmable AI Gateway library for Go, fully compatible with **OpenAI API** and **OpenResponses API** specifications.

## Features

- **Dual API Compatibility**: OpenAI API + [OpenResponses specification](https://www.openresponses.org/)
- **OpenResponses Endpoint**: `POST /v1/responses` with semantic streaming events
- **OpenAI Endpoints**: Chat Completions, Embeddings, Images
- **Real Streaming Support**: Server-Sent Events with OpenResponses semantic events
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
    openAI := provider.NewHTTPProviderWithBaseURLAndPath(
        "https://api.openai.com/v1",
        "your-api-key",
        "/v1",  // Strip /v1 from endpoint
    )

    // Configure models
    registry := model.NewMapModelRegistry()
    registry.Register("gpt-4", openAI)

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

### OpenResponses API

The gateway implements the [OpenResponses specification](https://www.openresponses.org/) with full streaming support:

```bash
# Non-streaming
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "input": "Hello, how are you?"
  }'

# Streaming
curl http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "input": "Count to 10",
    "stream": true
  }'
```

**OpenResponses Features:**
- Semantic streaming events (`response.created`, `response.in_progress`, `response.output_text.delta`, etc.)
- Message items with role-based content
- Tool calling support
- Proper error format with `type`, `message`, `param` fields

### OpenAI API (Compatible)

```bash
# Chat Completions
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Embeddings
curl http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello world",
    "model": "text-embedding-3-small"
  }'

# Images (DALL-E)
curl http://localhost:8080/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "model": "dall-e-3",
    "prompt": "a cat"
  }'
```

## Streaming

### OpenResponses Streaming

OpenResponses uses semantic events for streaming:

```
event: response.created
data: {"type":"response.created","response_id":"resp_abc123",...}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"Hello","sequence_number":1,...}

event: response.completed
data: {"type":"response.completed",...}

data: [DONE]
```

### OpenAI Streaming

Traditional OpenAI-style streaming is also supported:

```go
req := openai.ChatCompletionRequest{
    Model:    "gpt-4",
    Messages: messages,
    Stream:   true,
}
```

## OpenResponses Request Format

```json
{
  "model": "gpt-4",
  "input": "Your prompt here",
  "stream": false,
  "temperature": 0.7,
  "max_output_tokens": 1000,
  "tools": [...],
  "tool_choice": "auto",
  "truncation": "auto"
}
```

## OpenResponses Response Format

```json
{
  "id": "resp_abc123",
  "object": "response",
  "status": "completed",
  "created_at": 1234567890,
  "completed_at": 1234567895,
  "model": "gpt-4",
  "output": [
    {
      "id": "msg_xyz789",
      "type": "message",
      "status": "completed",
      "role": "assistant",
      "content": [
        {
          "type": "output_text",
          "text": "Response text here"
        }
      ]
    }
  ],
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20,
    "total_tokens": 30
  }
}
```

## Hook System

Hooks allow you to customize request/response processing:

### Authentication Hook

```go
type AuthHook struct{}

func (h *AuthHook) Name() string { return "auth" }

func (h *AuthHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
    // Validate API key, return (success, tenantID, error)
    return true, "user-id", nil
}

hooks := hook.NewRegistry()
hooks.Register(&AuthHook{})
```

### Request/Response Hooks

```go
type LoggingHook struct{}

func (h *LoggingHook) BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error {
    log.Printf("Request: %+v", req)
    return nil
}

func (h *LoggingHook) AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error {
    log.Printf("Response: %+v", resp)
    return nil
}
```

## Providers

### HTTP Provider (Generic)

```go
import "github.com/deeplooplabs/ai-gateway/provider"

// Standard OpenAI-compatible provider
provider := provider.NewHTTPProviderWithBaseURL(
    "https://api.openai.com/v1",
    "your-api-key",
)

// Provider with BasePath (for APIs that include /v1 in base URL)
provider := provider.NewHTTPProviderWithBaseURLAndPath(
    "https://api.siliconflow.cn/v1",  // BaseURL includes /v1
    "your-api-key",
    "/v1",  // Strip /v1 from endpoint to avoid duplication
)
```

### Configuration Options

```go
config := provider.NewProviderConfig("my-provider").
    WithBaseURL("https://api.example.com/v1").
    WithBasePath("/v1").  // Strip from endpoint
    WithAPIKey("your-key").
    WithAPIType(provider.APITypeAll).
    WithTimeout(30 * time.Second)

provider := provider.NewHTTPProvider(config)
```

## Model Registry

```go
registry := model.NewMapModelRegistry()

// Simple registration
registry.Register("gpt-4", provider)

// With options
registry.RegisterWithOptions("gpt-4", provider,
    model.WithModelRewrite("gpt-4-turbo"),      // Rewrite model name
    model.WithPreferredAPI(provider.APITypeChatCompletions),
)
```

## Docker Support

```bash
# Build
docker build -t ai-gateway-example -f example/Dockerfile .

# Run
docker run -p 8083:8083 \
  -e OPENAI_BASE_URL=https://api.openai.com/v1 \
  -e OPENAI_API_KEY=your-key \
  ai-gateway-example
```

See `example/` directory for a complete working example.

## License

MIT
