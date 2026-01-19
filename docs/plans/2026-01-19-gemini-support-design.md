# Gemini API Support Design

**Date**: 2026-01-19
**Status**: Design

## Overview

Add Google Gemini API support to the AI Gateway while maintaining the existing architecture. Support both Google Go SDK and direct HTTP approaches, with dual format compatibility (OpenAI and native Gemini formats).

## Requirements

1. **Dual Provider Implementation**: Both SDK (`cloud.google.com/go/ai`) and HTTP approaches
2. **Endpoints**: Chat/Completions, Embeddings, Images (Gemini doesn't support images natively)
3. **Dual Format**: Accept both OpenAI and native Gemini request formats
4. **Authentication**: API Key only (consistent with existing HTTPProvider)
5. **Interface**: Implement existing `Provider` interface without breaking changes

## Architecture

### Provider Abstraction with Format Detection

```
┌─────────────────────────────────────────────────────────────┐
│                     Model Registry                           │
│  gpt-4 → OpenAIProvider                                      │
│  gemini-pro → GeminiSDKProvider                             │
│  gemini-flash → GeminiHTTPProvider                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Provider Interface                      │
│  SendRequest(ctx, endpoint, req) (*Response, error)         │
│  SendRequestStream(ctx, endpoint, req) (<-chan Chunk, <-chan Error)│
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
    GeminiSDKProvider   GeminiHTTPProvider  HTTPProvider
         (SDK)              (HTTP)        (OpenAI)
            │                   │
            └─────────┬─────────┘
                      ▼
            Format Converter (OpenAI ↔ Gemini)
```

### Package Structure

```
ai-gateway/
├── gemini/
│   ├── types.go           # Native Gemini request/response types
│   ├── converter.go       # OpenAI ↔ Gemini format conversion
│   └── converter_test.go
├── provider/
│   ├── gemini_sdk.go      # SDK-based provider implementation
│   ├── gemini_sdk_test.go
│   ├── gemini_http.go     # HTTP-based provider implementation
│   └── gemini_http_test.go
```

## Format Converter

**Package: `gemini/converter.go`**

Handles bidirectional translation between OpenAI and Gemini formats.

### OpenAI → Gemini Mapping

| OpenAI Field | Gemini Field | Notes |
|--------------|--------------|-------|
| `messages` | `contents` | Array of content objects |
| `messages[].role` | `contents[].role` | user→user, assistant→model, system→system |
| `messages[].content` | `contents[].parts[].text` | Content becomes text part |
| `tools` | `tools` | Function declarations |
| `tool_calls` | `functionCalls` | Function call parts |
| `temperature` | `generationConfig.temperature` | Nested in config |
| `max_tokens` | `generationConfig.maxOutputTokens` | Nested in config |
| `stream: true` | Use `streamGenerateContent` | Different method |

### Gemini → OpenAI Mapping

| Gemini Field | OpenAI Field | Notes |
|--------------|--------------|-------|
| `candidates` | `choices` | Response alternatives |
| `candidates[].content.parts` | `choices[].message.content` | Extract text parts |
| `candidates[].finishReason` | `choices[].finish_reason` | Reason codes |
| `usageMetadata` | `usage` | Token counts |
| `content.parts` | `message.tool_calls` | Function calls |

### Native Gemini Types

```go
// gemini/types.go
type GenerateContentRequest struct {
    Contents          []Content          `json:"contents"`
    Tools             []Tool             `json:"tools,omitempty"`
    GenerationConfig  GenerationConfig   `json:"generationConfig,omitempty"`
}

type Content struct {
    Role  string `json:"role,omitempty"`  // "user", "model", "function"
    Parts []Part `json:"parts"`
}

type Part struct {
    Text         string                 `json:"text,omitempty"`
    FunctionCall map[string]interface{} `json:"functionCall,omitempty"`
    InlineData   *InlineData            `json:"inlineData,omitempty"`
}

type GenerateContentResponse struct {
    Candidates     []Candidate `json:"candidates"`
    UsageMetadata  Usage       `json:"usageMetadata"`
}

type Candidate struct {
    Content       Content `json:"content"`
    FinishReason  string  `json:"finishReason"`
    Index         int     `json:"index"`
}
```

## Provider Implementations

### GeminiSDKProvider

**Package: `provider/gemini_sdk.go`**

Uses the official Google AI SDK (`cloud.google.com/go/ai`).

```go
type GeminiSDKProvider struct {
    client *ai.Client
    model  string  // e.g., "gemini-2.0-flash-exp"
}

func NewGeminiSDKProvider(apiKey, model string) (*GeminiSDKProvider, error)
func (p *GeminiSDKProvider) Name() string { return "gemini-sdk" }
func (p *GeminiSDKProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
func (p *GeminiSDKProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
```

**Implementation logic:**
1. Detect request format (OpenAI vs native Gemini)
2. Convert to Gemini format if needed
3. Call SDK methods
4. Convert response back to OpenAI format
5. Return standard OpenAI response

### GeminiHTTPProvider

**Package: `provider/gemini_http.go`**

Makes direct HTTP requests to Gemini REST API.

```go
type GeminiHTTPProvider struct {
    BaseURL string  // "https://generativelanguage.googleapis.com/v1beta"
    APIKey  string
    Client  *http.Client
}

func NewGeminiHTTPProvider(apiKey string) *GeminiHTTPProvider
func (p *GeminiHTTPProvider) Name() string { return "gemini-http" }
func (p *GeminiHTTPProvider) SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
func (p *GeminiHTTPProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error)
```

**HTTP Endpoints:**
- Chat: `POST /v1beta/models/{model}:generateContent`
- Stream: `POST /v1beta/models/{model}:streamGenerateContent`
- Embeddings: `POST /v1beta/models/{model}:embedContent`
- Images: Not supported → return error

**Headers:**
```
Content-Type: application/json
x-goog-api-key: {API_KEY}
```

## Handler Integration

**No Handler Changes Required**

The existing `ChatHandler`, `EmbeddingsHandler`, and `ImagesHandler` already work with the `Provider` interface. They don't need modification.

**Gemini Images Handling:**

Since Gemini doesn't have a native image generation endpoint, when `GeminiHTTPProvider.SendRequest()` is called with the images endpoint:
- Returns error: `"image generation not supported for Gemini provider"`

## Usage Example

```go
import (
    "github.com/deeplooplabs/ai-gateway/gateway"
    "github.com/deeplooplabs/ai-gateway/model"
    "github.com/deeplooplabs/ai-gateway/provider"
)

// Setup providers
openAI := provider.NewHTTPProvider("https://api.openai.com", "sk-...")
geminiSDK, _ := provider.NewGeminiSDKProvider("AIza...", "gemini-2.0-flash-exp")
geminiHTTP := provider.NewGeminiHTTPProvider("AIza...")

// Register models
registry := model.NewMapModelRegistry()
registry.Register("gpt-4", openAI, "")
registry.Register("gemini-pro", geminiSDK, "")
registry.Register("gemini-flash", geminiHTTP, "gemini-2.0-flash-exp")

// Create gateway
gw := gateway.New(gateway.WithModelRegistry(registry))

// All routes work with all providers
// POST /v1/chat/completions with model="gemini-pro" → Gemini SDK
// POST /v1/chat/completions with model="gemini-flash" → Gemini HTTP
// POST /v1/embeddings with model="gemini-pro" → Converted to Gemini format
```

## Dependencies

```
require cloud.google.com/go/ai v0.2.0
```

## Error Handling

| Error Type | Format |
|------------|--------|
| Format conversion | `fmt.Errorf("convert request: %w", err)` |
| Unsupported endpoint | `fmt.Errorf("endpoint not supported: %s", endpoint)` |
| SDK errors | Wrapped with context |
| HTTP errors | Same pattern as HTTPProvider |

## Streaming

Gemini's SSE format differs from OpenAI. The converter translates Gemini chunks to OpenAI-compatible chunks before sending to the channel.

**Gemini SSE Format:**
```
data: {"candidates":[{"content":{"parts":[{"text":"..."}]}}

data: {"candidates":[{"content":{"parts":[{"text":"..."}],"finishReason":"STOP"}]}
```

**Translated to OpenAI Format:**
```
data: {"id":"...","choices":[{"delta":{"content":"..."},"index":0}],"object":"chat.completion.chunk"}

data: {"id":"...","choices":[{"delta":{},"finish_reason":"stop","index":0}],"object":"chat.completion.chunk"}
```

## Testing

1. **Unit Tests**: Converter round-trip (OpenAI ↔ Gemini)
2. **Mock Tests**: Both providers with mocked responses
3. **Integration Tests**: Optional (require API keys)

## Backward Compatibility

- Existing `Provider` interface unchanged
- All existing providers continue to work
- No changes to handlers or gateway
- New dependency is optional (only if using Gemini SDK)
