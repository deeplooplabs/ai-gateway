package e2e

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/deeplooplabs/ai-gateway/gateway"
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
	"github.com/deeplooplabs/ai-gateway/provider"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

// TestEnvironment provides a complete test setup with gateway, mock provider, and OpenAI client
type TestEnvironment struct {
	Server       *httptest.Server
	Client       *openai.Client
	Gateway      *gateway.Gateway
	Registry     *model.MapModelRegistry
	MockProvider *E2EMockProvider
	T            *testing.T
}

// NewTestEnvironment creates a new test environment with all necessary components
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	// Create mock provider
	mockProvider := NewE2EMockProvider()

	// Setup model registry with test models
	registry := model.NewMapModelRegistry()

	// Register chat models
	registry.RegisterWithOptions("gpt-4", mockProvider,
		model.WithPreferredAPI(provider.APITypeChatCompletions),
	)
	registry.RegisterWithOptions("gpt-3.5-turbo", mockProvider,
		model.WithPreferredAPI(provider.APITypeChatCompletions),
	)

	// Register embedding models
	registry.RegisterWithOptions("text-embedding-3-small", mockProvider,
		model.WithPreferredAPI(provider.APITypeEmbeddings),
	)
	registry.RegisterWithOptions("text-embedding-3-large", mockProvider,
		model.WithPreferredAPI(provider.APITypeEmbeddings),
	)

	// Register image models
	registry.RegisterWithOptions("dall-e-3", mockProvider,
		model.WithPreferredAPI(provider.APITypeImages),
	)

	// Create hooks (no auth hook by default)
	hooks := hook.NewRegistry()

	// Create gateway
	gw := gateway.New(
		gateway.WithModelRegistry(registry),
		gateway.WithHooks(hooks),
	)

	// Start test server
	server := httptest.NewServer(gw)

	// Create OpenAI client pointing to test server
	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = server.URL + "/v1"
	client := openai.NewClientWithConfig(config)

	env := &TestEnvironment{
		Server:       server,
		Client:       client,
		Gateway:      gw,
		Registry:     registry,
		MockProvider: mockProvider,
		T:            t,
	}

	// Cleanup on test completion
	t.Cleanup(func() {
		server.Close()
	})

	return env
}

// NewTestEnvironmentWithAuth creates a test environment with authentication enabled
func NewTestEnvironmentWithAuth(t *testing.T, validKey string) *TestEnvironment {
	// Create mock provider
	mockProvider := NewE2EMockProvider()

	// Setup model registry
	registry := model.NewMapModelRegistry()
	registry.RegisterWithOptions("gpt-4", mockProvider,
		model.WithPreferredAPI(provider.APITypeChatCompletions),
	)

	// Create hooks with auth hook
	hooks := hook.NewRegistry()
	hooks.Register(&TestAuthHook{ValidKey: validKey})

	// Create gateway
	gw := gateway.New(
		gateway.WithModelRegistry(registry),
		gateway.WithHooks(hooks),
	)

	// Start test server
	server := httptest.NewServer(gw)

	// Create OpenAI client with auth token
	config := openai.DefaultConfig(validKey)
	config.BaseURL = server.URL + "/v1"
	client := openai.NewClientWithConfig(config)

	env := &TestEnvironment{
		Server:       server,
		Client:       client,
		Gateway:      gw,
		Registry:     registry,
		MockProvider: mockProvider,
		T:            t,
	}

	t.Cleanup(func() {
		server.Close()
	})

	return env
}

// TestAuthHook is a simple authentication hook for testing
type TestAuthHook struct {
	ValidKey string
}

func (h *TestAuthHook) Name() string {
	return "test-auth"
}

func (h *TestAuthHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
	// Strip "Bearer " prefix if present
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	if apiKey == h.ValidKey {
		return true, "test-tenant", nil
	}
	return false, "", nil
}

var _ hook.AuthenticationHook = (*TestAuthHook)(nil)

// ValidateStreamingResponse accumulates and validates streaming response
func ValidateStreamingResponse(t *testing.T, stream *openai.ChatCompletionStream) string {
	t.Helper()

	var accumulated strings.Builder
	var finishReason string
	chunkCount := 0

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "stream should not error")

		chunkCount++

		// Validate chunk structure
		require.NotEmpty(t, chunk.ID, "chunk should have ID")
		require.Equal(t, "chat.completion.chunk", chunk.Object, "chunk object should be chat.completion.chunk")
		require.NotEmpty(t, chunk.Model, "chunk should have model")
		require.NotEmpty(t, chunk.Choices, "chunk should have choices")

		// Accumulate content
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			accumulated.WriteString(choice.Delta.Content)

			if choice.FinishReason != "" {
				finishReason = string(choice.FinishReason)
			}
		}
	}

	require.Greater(t, chunkCount, 0, "should receive at least one chunk")
	require.NotEmpty(t, finishReason, "should receive finish_reason")

	return accumulated.String()
}
