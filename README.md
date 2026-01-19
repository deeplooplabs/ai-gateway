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

Streaming responses are sent using Server-Sent Events (SSE) format.

## Hook System

Hooks allow you to customize request/response processing:

### Authentication Hook

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

### Streaming Hook

```go
type LoggingHook struct{}

func (h *LoggingHook) Name() string { return "logger" }

func (h *LoggingHook) OnChunk(ctx context.Context, chunk []byte) ([]byte, error) {
    log.Printf("Chunk: %s", string(chunk))
    return chunk, nil  // Return modified chunk
}
```

## Providers

### OpenAI Provider

```go
import (
    "github.com/deeplooplabs/ai-gateway/provider"
)

openAI := provider.NewHTTPProvider("https://api.openai.com", "your-api-key")
registry.Register("gpt-4", openAI, "")
```

### Gemini Provider

```go
import (
    "github.com/deeplooplabs/ai-gateway/provider"
)

// HTTP-based Gemini provider
geminiHTTP := provider.NewGeminiHTTPProvider("your-api-key")
registry.Register("gemini-pro", geminiHTTP, "gemini-2.0-flash-exp")
```

## License

MIT
