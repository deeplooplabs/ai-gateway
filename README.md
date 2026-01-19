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
