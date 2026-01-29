# E2E Tests for AI Gateway

This directory contains comprehensive End-to-End (E2E) tests for the AI Gateway using the official OpenAI Go client library.

## Overview

The E2E tests verify complete API compatibility with OpenAI endpoints by using the real OpenAI Go client (`github.com/sashabaranov/go-openai`) against a test server running the AI Gateway. This ensures that the gateway is fully compatible with OpenAI's API specification.

## Test Coverage

### Total: 24 Tests

#### Chat Completions (7 tests)
- ✅ `TestE2E_ChatCompletions_Basic` - Basic chat completion
- ✅ `TestE2E_ChatCompletions_WithOptions` - With temperature, max_tokens
- ✅ `TestE2E_ChatCompletions_MultipleMessages` - Conversation history
- ✅ `TestE2E_ChatCompletions_ModelNotFound` - Error handling
- ✅ `TestE2E_ChatCompletions_Streaming` - Streaming responses
- ✅ `TestE2E_ChatCompletions_StreamingFinishReason` - Stream finish reason
- ✅ `TestE2E_ChatCompletions_StreamingContextCancel` - Context cancellation

#### Embeddings (3 tests)
- ✅ `TestE2E_Embeddings_SingleInput` - Single text embedding
- ✅ `TestE2E_Embeddings_MultipleInputs` - Multiple text embeddings
- ✅ `TestE2E_Embeddings_WithDimensions` - Custom dimensions

#### Images (3 tests)
- ✅ `TestE2E_Images_Basic` - Basic image generation
- ✅ `TestE2E_Images_MultipleImages` - Multiple images (N=2)
- ✅ `TestE2E_Images_WithOptions` - With size, quality, style

#### Models (1 test)
- ✅ `TestE2E_Models_List` - List available models

#### Error Handling (4 tests)
- ✅ `TestE2E_ErrorHandling_InvalidJSON` - Malformed requests
- ✅ `TestE2E_ErrorHandling_MissingRequiredField` - Validation errors
- ✅ `TestE2E_ErrorHandling_ProviderError` - Provider errors
- ✅ `TestE2E_ErrorHandling_RateLimitError` - Rate limit errors

#### Authentication (2 tests)
- ✅ `TestE2E_Authentication_ValidAPIKey` - Successful auth
- ✅ `TestE2E_Authentication_InvalidAPIKey` - Failed auth

#### OpenResponses (4 tests)
- ✅ `TestE2E_OpenResponses_Basic` - Non-streaming request
- ✅ `TestE2E_OpenResponses_Streaming` - Streaming SSE events
- ✅ `TestE2E_OpenResponses_WithTools` - Tool/function calling
- ✅ `TestE2E_OpenResponses_ErrorResponse` - Error format

## Architecture

### Files

| File | Purpose | Lines |
|------|---------|-------|
| `client_test.go` | OpenAI client E2E tests (20 tests) | ~750 |
| `openresponses_test.go` | OpenResponses endpoint tests (4 tests) | ~300 |
| `mock_provider.go` | Unified mock provider implementation | ~270 |
| `helpers.go` | Test environment setup & utilities | ~180 |

### Test Environment

Each test creates an isolated environment with:

1. **Mock Provider** - Implements `provider.Provider` interface
2. **Model Registry** - Registers test models (gpt-4, text-embedding-3-small, dall-e-3, etc.)
3. **Gateway** - Fresh gateway instance with hooks
4. **Test Server** - `httptest.NewServer` for in-memory HTTP testing
5. **OpenAI Client** - Official client configured to point to test server

```go
env := NewTestEnvironment(t)
defer env.Server.Close() // Automatic cleanup

// Use official OpenAI client
resp, err := env.Client.CreateChatCompletion(ctx, request)
```

### Mock Provider

The `E2EMockProvider` is a unified mock that:

- Supports all API types (chat, embeddings, images)
- Handles both streaming and non-streaming requests
- Thread-safe with mutex-protected configuration
- Configurable responses per test
- Realistic streaming with delays (10ms between chunks)

```go
// Configure mock response
env.MockProvider.SetChatResponse(&openai.ChatCompletionResponse{
    ID: "test-123",
    Choices: []openai.Choice{{Message: openai.Message{Content: "Hello"}}},
})

// Configure error simulation
env.MockProvider.SetError(true, "simulated error")
```

## Running Tests

### Run all E2E tests
```bash
go test ./e2e/... -v
```

### Run specific test
```bash
go test ./e2e/... -v -run TestE2E_ChatCompletions_Streaming
```

### Run with coverage
```bash
go test ./e2e/... -v -coverprofile=coverage.out
```

### Run with race detector
```bash
go test ./e2e/... -v -race
```

### Skip in short mode
```bash
go test ./e2e/... -short  # Skips E2E tests
```

## Test Strategy

### Validation Approach

**For Non-Streaming Tests:**
1. Send request via OpenAI client
2. Verify HTTP 200 status
3. Validate response structure matches OpenAI spec
4. Check required fields (id, object, model, choices/data, usage)
5. Verify content is not empty

**For Streaming Tests:**
1. Create stream via OpenAI client
2. Iterate through chunks with `stream.Recv()`
3. Validate each chunk structure
4. Accumulate delta content
5. Verify finish_reason in last chunk
6. Verify EOF received

**For OpenResponses Tests:**
1. Use raw HTTP client
2. Parse SSE events manually
3. Verify event sequence (created → in_progress → output → completed)
4. Validate [DONE] marker
5. Check event types and sequence numbers

### Verification Utilities

```go
// Validate streaming response
content := ValidateStreamingResponse(t, stream)
assert.Contains(t, content, "expected text")

// Parse SSE events
events := parseSSEEvents(t, resp.Body)
assert.True(t, eventTypes["response.created"])
```

## Dependencies

- `github.com/sashabaranov/go-openai` v1.41.2+ - Official OpenAI Go client
- `github.com/stretchr/testify` v1.8.4+ - Assertions and test utilities

## Design Principles

1. **Isolated Tests** - Each test gets fresh server instance (no shared state)
2. **Real Client** - Uses official OpenAI client library (not mocks)
3. **Fast Execution** - In-memory httptest servers (~0.8s total)
4. **Thread-Safe** - Mock provider uses mutexes for concurrent safety
5. **Comprehensive** - Covers success paths, errors, streaming, auth
6. **Maintainable** - Clear test structure, reusable helpers

## Coverage

Current coverage: **89.6%** of statements in e2e package

Coverage includes:
- All API endpoint types
- Streaming and non-streaming paths
- Error handling scenarios
- Authentication flows
- OpenAI and OpenResponses formats

## Future Enhancements

Potential additions (not currently implemented):

- Function/tool calling tests (advanced scenarios)
- Image upload endpoints (variations, edits)
- Audio endpoints (transcription, TTS)
- Performance/load testing
- Real provider integration tests

## CI Integration

Tests can be excluded from short test runs:

```go
if testing.Short() {
    t.Skip("skipping E2E test in short mode")
}
```

CI configuration example:
```yaml
# Fast unit tests
- run: go test ./... -short

# Full test suite including E2E
- run: go test ./... -v -timeout=5m
```

## Contributing

When adding new tests:

1. Follow naming convention: `TestE2E_{Category}_{TestName}`
2. Add skip for short mode: `if testing.Short() { t.Skip() }`
3. Use `NewTestEnvironment(t)` for setup
4. Add test to this README's coverage section
5. Ensure test is isolated (no shared state)
6. Use descriptive assertions with context

## Troubleshooting

**Tests hanging?**
- Check for missing `defer stream.Close()`
- Verify context cancellation in streaming tests
- Look for deadlocks in mock provider

**Flaky tests?**
- Ensure proper cleanup with `t.Cleanup()`
- Check for race conditions (run with `-race`)
- Verify mock provider is reset between tests

**Import errors?**
- Run `go mod tidy` to sync dependencies
- Check that OpenAI client version is compatible
