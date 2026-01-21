# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**DeepLoop AI Gateway** is a programmable AI Gateway library written in Go that provides OpenAI API compatibility and aims to comply with the [OpenResponses specification](https://www.openresponses.org/). It is designed as an **embeddable library** (not a standalone service) that can be integrated into Go applications to provide unified access to multiple LLM providers.

## Development Commands

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./handler/...
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
│  - ChatHandler                          │
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
```

### Key Packages

- **`gateway/`**: Main HTTP handler implementing `http.Handler`. Routes requests to appropriate handlers. Options pattern for configuration.
- **`handler/`**: HTTP handlers for each API endpoint (chat, embeddings, images). Handles authentication hooks, request parsing, and response assembly.
- **`hook/`**: Extensible hook system with 4 hook types. Hooks are called at specific points in the request lifecycle.
- **`model/`**: Model registry that maps model names to providers with optional model name rewriting (for provider-specific model names).
- **`provider/`**: Abstract interface for LLM providers. Includes HTTP provider for REST APIs and Gemini-specific provider with type conversion.

### OpenAI Types

All OpenAI request/response types are defined in `provider/openai/types.go`. This is the canonical location for API schemas.

### Streaming Implementation

Streaming uses Server-Sent Events (SSE). The provider returns a channel of `StreamChunk` which the handler processes and writes to the response in SSE format. Each chunk can be modified by `StreamingHook` implementations before being written.

## Hook System Usage

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

## Adding a New Provider

1. Implement the `provider.Provider` interface in `provider/`
2. Optionally implement streaming by returning channels for `StreamChunk` and errors
3. Register models with the provider in the model registry

## OpenResponses Specification Compliance

This project aims to comply with the [OpenResponses specification](https://www.openresponses.org/specification). Reference: [OpenResponses GitHub](https://github.com/openresponses/openresponses), [OpenAPI Schema](https://www.openresponses.org/reference).

### API Endpoint

OpenResponses uses `POST /v1/responses` as the primary endpoint. This gateway currently implements OpenAI-compatible endpoints (`/v1/chat/completions`, `/v1/embeddings`, `/v1/images/generations`). Future work should consider adding the `/v1/responses` endpoint.

### Request Format

OpenResponses requests use these key parameters:

| Parameter | Description |
|-----------|-------------|
| `model` | The model to use (e.g., 'gpt-4o') |
| `input` | Context as string or array of items (UserMessageItemParam, SystemMessageItemParam, etc.) |
| `previous_response_id` | ID of previous response for continuation |
| `tools` | Array of available tools (FunctionToolParam) |
| `tool_choice` | Controls tool usage: `none`, `auto`, `required`, or specific function |
| `stream` | Enable SSE streaming |
| `temperature`, `top_p` | Sampling parameters |
| `max_output_tokens` | Maximum output tokens |
| `truncation` | `auto` or `disabled` - context truncation behavior |
| `instructions` | Additional instructions for the model |

### Response Format

OpenResponses responses include:

| Field | Description |
|-------|-------------|
| `id` | Unique response ID |
| `object` | Always "response" |
| `status` | `in_progress`, `completed`, `failed`, `incomplete` |
| `output` | Array of output items (Message, FunctionCall, Reasoning, etc.) |
| `error` | Error object if failed |
| `usage` | Token usage statistics |
| `created_at`, `completed_at` | Unix timestamps |

### Item Types

OpenResponses defines these item types:

| Type | Description |
|------|-------------|
| `message` | User/assistant/system/developer messages |
| `function_call` | Function tool calls generated by model |
| `function_call_output` | Results from function calls |
| `reasoning` | Model's internal reasoning process |
| Custom | Provider-specific types prefixed with slug (e.g., `openai:web_search_call`) |

### Content Types

**UserContent** (input to model):
- `input_text` - Text input
- `input_image` - Image input (URL or base64)
- `input_file` - File input

**ModelContent** (output from model):
- `output_text` - Text output with optional annotations/logprobs
- `refusal` - Model refusal response

### Streaming Events

OpenResponses streaming uses semantic events, not raw deltas. Event types include:

**State Machine Events:**
- `response.created` - Response initialized
- `response.in_progress` - Response is being generated
- `response.completed` - Response finished successfully
- `response.failed` - Response encountered an error
- `response.incomplete` - Response incomplete (token budget exhausted)

**Item Events:**
- `response.output_item.added` - New output item started
- `response.output_item.done` - Item completed

**Content Events:**
- `response.content_part.added` - New content part started
- `response.output_text.delta` - Text delta
- `response.output_text.done` - Text part completed
- `response.function_call_arguments.delta` - Function call delta
- `response.function_call_arguments.done` - Function call completed

**Reasoning Events:**
- `response.reasoning.delta` - Reasoning content delta
- `response.reasoning.done` - Reasoning completed
- `response.reasoning_summary.delta` - Summary delta
- `response.reasoning_summary.done` - Summary completed

### Streaming Protocol Requirements

- **MUST** use `Content-Type: text/event-stream`
- **MUST** terminate with literal string `[DONE]`
- **MUST** use `event` field matching `type` in event body
- **SHOULD NOT** use `id` field in SSE events
- Each event MUST include `sequence_number` for ordering

Example streaming event:
```
event: response.output_text.delta
data: {"type":"response.output_text.delta","sequence_number":10,"item_id":"msg_abc123","output_index":0,"content_index":0,"delta":"Hello"}
```

### Error Response Format

Errors follow the OpenResponses schema:

```json
{
  "error": {
    "message": "Human-readable error description",
    "type": "invalid_request_error|server_error|not_found|model_error|too_many_requests",
    "code": "optional_specific_code",
    "param": "optional_parameter_name"
  }
}
```

Error types and HTTP status codes:
| Type | Status Code | Description |
|------|-------------|-------------|
| `invalid_request_error` | 400 | Malformed or semantically invalid request |
| `not_found` | 404 | Resource does not exist |
| `too_many_requests` | 429 | Rate limit exceeded |
| `server_error` | 500 | Internal server failure |
| `model_error` | 500 | Model failed during processing |

Current implementation in `handler/*.go` defines:
- `invalid_request_error` (400)
- `not_found_error` (404)
- `api_error` (502) - should map to `server_error`

### Extension Support

Provider-specific extensions MUST follow naming conventions:
- Item types: `{provider_slug}:{item_type}` (e.g., `openai:web_search_call`)
- Streaming events: `{provider_slug}:{event_type}` (e.g., `acme:trace_event`)

### Migration Path

The current implementation uses OpenAI-compatible formats. To achieve full OpenResponses compliance:

1. Add `/v1/responses` endpoint alongside existing OpenAI endpoints
2. Implement OpenResponses request/response types in new package (e.g., `openresponses/`)
3. Update streaming to use semantic events instead of OpenAI chunks
4. Add support for `previous_response_id` for conversation continuation
5. Implement reasoning item support
6. Add proper tool_choice modes including `allowed_tools`

## Testing Strategy

- Unit tests for each package alongside source files (`*_test.go`)
- Handler tests mock provider responses
- Provider tests use recorded responses or test servers
- Hook tests verify execution order and context propagation

## Configuration

The gateway uses functional options for configuration:

```go
gw := gateway.New(
    gateway.WithModelRegistry(registry),
    gateway.WithHooks(hooks),
)
```

All available options are in `gateway/option.go`.
