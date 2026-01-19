# AI Gateway API Expansion Design

**Date**: 2026-01-19
**Status**: Design

## Overview

Extend the AI Gateway library with additional OpenAI API endpoints (Embeddings, Images) and implement real streaming response support using a simplified channel-based approach.

## Goals

1. Add support for OpenAI Embeddings and Images (DALL-E) endpoints
2. Implement real streaming response handling (replacing mock implementation)
3. Maintain backward compatibility with existing code

## Architecture

### New API Endpoints

#### 1. Embeddings Endpoint

**Path**: `/v1/embeddings`
**Method**: `POST`
**Purpose**: Convert text to vector embeddings for semantic search

**Request**:
```go
type EmbeddingRequest struct {
    Input          any    `json:"input"`           // string or []string
    Model          string `json:"model"`
    EncodingFormat string `json:"encoding_format,omitempty"` // "float" or "base64"
}
```

**Response**:
```go
type EmbeddingResponse struct {
    Object string      `json:"object"`
    Data   []Embedding `json:"data"`
    Model  string      `json:"model"`
    Usage  Usage       `json:"usage"`
}

type Embedding struct {
    Object    string    `json:"object"`
    Embedding []float32 `json:"embedding"`
    Index     int       `json:"index"`
}
```

#### 2. Images Endpoint

**Path**: `/v1/images/generations`
**Method**: `POST`
**Purpose**: Generate images using DALL-E

**Request**:
```go
type ImageRequest struct {
    Model   string `json:"model"`
    Prompt  string `json:"prompt"`
    N       int    `json:"n,omitempty"`
    Size    string `json:"size,omitempty"`    // "256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"
    Quality string `json:"quality,omitempty"` // "standard" or "hd"
    Style   string `json:"style,omitempty"`   // "vivid" or "natural"
}
```

**Response**:
```go
type ImageResponse struct {
    Created int64   `json:"created"`
    Data    []Image `json:"data"`
}

type Image struct {
    URL         string `json:"url,omitempty"`         // For DALL-E 2
    B64JSON     string `json:"b64_json,omitempty"`    // For DALL-E 3
    RevisedPrompt string `json:"revised_prompt,omitempty"`
}
```

### Streaming Response Implementation

#### Provider Interface Extension

```go
type Provider interface {
    Name() string
    SendRequest(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
    SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan StreamChunk, <-chan error)
}

type StreamChunk struct {
    Data []byte  // Raw SSE data line
    Done bool    // Stream end marker
}
```

#### SSE Parser

A lightweight SSE parser in `openai/sse.go`:

```go
type SSEParser struct {
    scanner *bufio.Scanner
}

func NewSSEReader(r io.Reader) *SSEReader
func (r *SSEReader) NextChunk() (event, data string, err error)
func (r *SSEReader) IsDone(data []byte) bool
```

**SSE Format**:
```
event: message
data: {"id":"...","choices":[...],"object":"chat.completion.chunk"}

data: [DONE]
```

#### Chat Handler Streaming Flow

```
1. Parse request → detect req.Stream == true
2. Resolve provider
3. BeforeRequest hooks
4. provider.SendRequestStream() → returns channels
5. Loop:
   a. Read chunk from channel
   b. Parse SSE line (extract data: {...})
   c. OnChunk hook (can modify chunk)
   d. Write to ResponseWriter
   e. Flush()
6. [DONE] or error → Close stream
```

#### StreamingHook Enhancement

```go
type StreamingHook interface {
    Hook
    OnChunk(ctx *Context, chunk []byte) ([]byte, error)  // Returns modified chunk
}
```

### HTTPProvider Streaming

```go
func (p *HTTPProvider) SendRequestStream(ctx context.Context, endpoint string, req *openai.ChatCompletionRequest) (<-chan StreamChunk, <-chan error) {
    chunkChan := make(chan StreamChunk, 16)
    errChan := make(chan error, 1)

    go func() {
        defer close(chunkChan)
        defer close(errChan)

        // Create HTTP request
        // Set headers including Authorization

        // Send request
        resp, err := p.Client.Do(httpReq)
        if err != nil {
            errChan <- err
            return
        }
        defer resp.Body.Close()

        // Read SSE line by line
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Bytes()

            // Check for [DONE]
            if bytes.HasPrefix(line, []byte("data: [DONE]")) {
                chunkChan <- StreamChunk{Done: true}
                return
            }

            // Extract data: content
            if bytes.HasPrefix(line, []byte("data:")) {
                data := bytes.TrimPrefix(line, []byte("data: "))
                chunkChan <- StreamChunk{Data: data}
            }
        }

        if err := scanner.Err(); err != nil {
            errChan <- err
        }
    }()

    return chunkChan, errChan
}
```

### Package Structure

```
ai-gateway/
├── openai/
│   ├── types.go           # Extended with Embeddings/Image types
│   ├── sse.go             # New: SSE parser
│   ├── types_test.go
│   └── sse_test.go        # New
├── provider/
│   ├── provider.go        # Extended: SendRequestStream
│   ├── provider_test.go
│   ├── http.go            # Extended: streaming implementation
│   └── http_test.go
├── handler/
│   ├── chat.go            # Modified: uses streaming provider
│   ├── chat_test.go
│   ├── embeddings.go      # New
│   ├── embeddings_test.go # New
│   ├── images.go          # New
│   └── images_test.go     # New
├── hook/
│   ├── hook.go            # Enhanced: StreamingHook returns modified chunk
│   └── hook_test.go
├── model/
│   ├── registry.go         # Supports model-specific streaming config
│   └── registry_test.go
├── gateway/
│   ├── gateway.go         # Modified: registers new routes
│   ├── gateway_test.go
│   └── option.go
├── context.go
├── error.go
└── README.md              # Updated with new endpoints
```

## Implementation Tasks

1. **SSE Parser** (`openai/sse.go`)
   - Parse SSE event/data lines
   - Detect [DONE] marker
   - Handle errors

2. **Provider Streaming** (`provider/`)
   - Add `SendRequestStream` to Provider interface
   - Implement streaming in HTTPProvider
   - Tests for streaming behavior

3. **Embeddings Handler** (`handler/embeddings.go`)
   - Parse EmbeddingRequest
   - Call Provider
   - Return EmbeddingResponse

4. **Images Handler** (`handler/images.go`)
   - Parse ImageRequest
   - Call Provider
   - Return ImageResponse (URL or base64)

5. **Chat Handler Streaming** (`handler/chat.go`)
   - Replace mock streaming with real Provider streaming
   - Integrate SSE parser
   - OnChunk hook support

6. **Gateway Routes** (`gateway/gateway.go`)
   - Register `/v1/embeddings`
   - Register `/v1/images/generations`

7. **Documentation**
   - Update README with new endpoints
   - Add streaming examples

## Backward Compatibility

- Existing non-streaming code unchanged
- Provider interface extends (not breaks)
- All existing tests continue to pass
