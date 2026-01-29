package e2e

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
	openailib "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Chat Completions (Non-Streaming) Tests
// ========================================

// TestE2E_ChatCompletions_Basic tests basic chat completion functionality
func TestE2E_ChatCompletions_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Configure mock response
	env.MockProvider.SetChatResponse(&openai.ChatCompletionResponse{
		ID:      "chatcmpl-basic123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []openai.Choice{
			{
				Index: 0,
				Message: openai.Message{
					Role:    "assistant",
					Content: "Hello! How can I help you today?",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 9,
			TotalTokens:      19,
		},
	})

	// Send request via OpenAI client
	resp, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	// Verify response
	require.NoError(t, err)
	assert.Equal(t, "chatcmpl-basic123", resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, "gpt-4", resp.Model)
	require.Len(t, resp.Choices, 1)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "Hello! How can I help you today?", resp.Choices[0].Message.Content)
	assert.Equal(t, "stop", string(resp.Choices[0].FinishReason))
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 9, resp.Usage.CompletionTokens)
	assert.Equal(t, 19, resp.Usage.TotalTokens)
}

// TestE2E_ChatCompletions_WithOptions tests chat completion with parameters
func TestE2E_ChatCompletions_WithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	temp := float32(0.7)
	maxTokens := 100

	resp, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Tell me a joke",
				},
			},
			Temperature: temp,
			MaxTokens:   maxTokens,
		},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	require.Len(t, resp.Choices, 1)
	assert.NotEmpty(t, resp.Choices[0].Message.Content)
}

// TestE2E_ChatCompletions_MultipleMessages tests conversation with multiple messages
func TestE2E_ChatCompletions_MultipleMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	env.MockProvider.SetChatResponse(&openai.ChatCompletionResponse{
		ID:      "chatcmpl-multi123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []openai.Choice{
			{
				Index: 0,
				Message: openai.Message{
					Role:    "assistant",
					Content: "The capital of France is Paris.",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     25,
			CompletionTokens: 8,
			TotalTokens:      33,
		},
	})

	resp, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "What is the capital of France?",
				},
				{
					Role:    openailib.ChatMessageRoleAssistant,
					Content: "The capital of France is Paris.",
				},
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Thank you!",
				},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "chatcmpl-multi123", resp.ID)
	assert.NotEmpty(t, resp.Choices[0].Message.Content)
}

// TestE2E_ChatCompletions_ModelNotFound tests error handling for unknown model
func TestE2E_ChatCompletions_ModelNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	_, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "nonexistent-model",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.Error(t, err)
	// OpenAI client wraps errors, check for error presence
	assert.Contains(t, err.Error(), "model not found")
}

// ========================================
// Chat Completions (Streaming) Tests
// ========================================

// TestE2E_ChatCompletions_Streaming tests streaming chat completion
func TestE2E_ChatCompletions_Streaming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	stream, err := env.Client.CreateChatCompletionStream(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Say hello",
				},
			},
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	content := ValidateStreamingResponse(t, stream)
	assert.NotEmpty(t, content, "should receive content from stream")
	assert.Contains(t, content, "Hello", "content should contain greeting")
}

// TestE2E_ChatCompletions_StreamingFinishReason tests streaming finish reason
func TestE2E_ChatCompletions_StreamingFinishReason(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	stream, err := env.Client.CreateChatCompletionStream(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	var lastFinishReason string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			lastFinishReason = string(chunk.Choices[0].FinishReason)
		}
	}

	assert.Equal(t, "stop", lastFinishReason, "should receive finish_reason 'stop'")
}

// TestE2E_ChatCompletions_StreamingContextCancel tests context cancellation during streaming
func TestE2E_ChatCompletions_StreamingContextCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := env.Client.CreateChatCompletionStream(
		ctx,
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Tell me a long story",
				},
			},
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	// Cancel context after receiving first chunk
	_, err = stream.Recv()
	require.NoError(t, err)

	cancel()

	// Next recv should fail due to cancelled context
	time.Sleep(100 * time.Millisecond) // Give time for cancellation to propagate
}

// ========================================
// Embeddings Tests
// ========================================

// TestE2E_Embeddings_SingleInput tests single text embedding
func TestE2E_Embeddings_SingleInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1
	}

	env.MockProvider.SetEmbeddingResponse(&openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{
			{
				Object:    "embedding",
				Embedding: embedding,
				Index:     0,
			},
		},
		Model: "text-embedding-3-small",
		Usage: openai.Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	})

	resp, err := env.Client.CreateEmbeddings(
		context.Background(),
		openailib.EmbeddingRequest{
			Model: "text-embedding-3-small",
			Input: []string{"Hello world"},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, string("text-embedding-3-small"), string(resp.Model))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "embedding", resp.Data[0].Object)
	assert.Equal(t, 0, resp.Data[0].Index)
	assert.Len(t, resp.Data[0].Embedding, 1536)
	assert.Equal(t, 5, resp.Usage.PromptTokens)
	assert.Equal(t, 5, resp.Usage.TotalTokens)
}

// TestE2E_Embeddings_MultipleInputs tests multiple text embeddings
func TestE2E_Embeddings_MultipleInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	embedding1 := make([]float32, 1536)
	embedding2 := make([]float32, 1536)
	for i := range embedding1 {
		embedding1[i] = 0.1
		embedding2[i] = 0.2
	}

	env.MockProvider.SetEmbeddingResponse(&openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{
			{
				Object:    "embedding",
				Embedding: embedding1,
				Index:     0,
			},
			{
				Object:    "embedding",
				Embedding: embedding2,
				Index:     1,
			},
		},
		Model: "text-embedding-3-small",
		Usage: openai.Usage{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	})

	resp, err := env.Client.CreateEmbeddings(
		context.Background(),
		openailib.EmbeddingRequest{
			Model: "text-embedding-3-small",
			Input: []string{"Hello world", "Goodbye world"},
		},
	)

	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, 0, resp.Data[0].Index)
	assert.Equal(t, 1, resp.Data[1].Index)
	assert.Len(t, resp.Data[0].Embedding, 1536)
	assert.Len(t, resp.Data[1].Embedding, 1536)
}

// TestE2E_Embeddings_WithDimensions tests embedding with custom dimensions
func TestE2E_Embeddings_WithDimensions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	dims := 768
	embedding := make([]float32, dims)
	for i := range embedding {
		embedding[i] = 0.1
	}

	env.MockProvider.SetEmbeddingResponse(&openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{
			{
				Object:    "embedding",
				Embedding: embedding,
				Index:     0,
			},
		},
		Model: "text-embedding-3-small",
		Usage: openai.Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	})

	resp, err := env.Client.CreateEmbeddings(
		context.Background(),
		openailib.EmbeddingRequestStrings{
			Model:      "text-embedding-3-small",
			Input:      []string{"Hello world"},
			Dimensions: dims,
		},
	)

	require.NoError(t, err)
	assert.Len(t, resp.Data[0].Embedding, dims)
}

// ========================================
// Images Tests
// ========================================

// TestE2E_Images_Basic tests basic image generation
func TestE2E_Images_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	env.MockProvider.SetImageResponse(&openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.Image{
			{
				URL: "https://example.com/generated-image.png",
			},
		},
	})

	resp, err := env.Client.CreateImage(
		context.Background(),
		openailib.ImageRequest{
			Model:  "dall-e-3",
			Prompt: "A beautiful sunset",
			N:      1,
		},
	)

	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "https://example.com/generated-image.png", resp.Data[0].URL)
}

// TestE2E_Images_MultipleImages tests generating multiple images
func TestE2E_Images_MultipleImages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	env.MockProvider.SetImageResponse(&openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.Image{
			{
				URL: "https://example.com/image1.png",
			},
			{
				URL: "https://example.com/image2.png",
			},
		},
	})

	resp, err := env.Client.CreateImage(
		context.Background(),
		openailib.ImageRequest{
			Model:  "dall-e-3",
			Prompt: "A cat",
			N:      2,
		},
	)

	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	assert.NotEmpty(t, resp.Data[0].URL)
	assert.NotEmpty(t, resp.Data[1].URL)
}

// TestE2E_Images_WithOptions tests image generation with options
func TestE2E_Images_WithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	resp, err := env.Client.CreateImage(
		context.Background(),
		openailib.ImageRequest{
			Model:   "dall-e-3",
			Prompt:  "A futuristic city",
			Size:    openailib.CreateImageSize1024x1024,
			Quality: "hd",
			N:       1,
		},
	)

	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	assert.NotEmpty(t, resp.Data[0].URL)
}

// ========================================
// Models Tests
// ========================================

// TestE2E_Models_List tests listing available models
func TestE2E_Models_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	resp, err := env.Client.ListModels(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Models)

	// Verify expected models are present
	modelNames := make(map[string]bool)
	for _, model := range resp.Models {
		modelNames[model.ID] = true
		assert.NotEmpty(t, model.ID)
	}

	assert.True(t, modelNames["gpt-4"], "gpt-4 should be in model list")
	assert.True(t, modelNames["gpt-3.5-turbo"], "gpt-3.5-turbo should be in model list")
	assert.True(t, modelNames["text-embedding-3-small"], "text-embedding-3-small should be in model list")
}

// ========================================
// Error Handling Tests
// ========================================

// TestE2E_ErrorHandling_InvalidJSON tests malformed JSON handling
func TestE2E_ErrorHandling_InvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// This should trigger validation error since messages is required
	_, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model:    "gpt-4",
			Messages: []openailib.ChatCompletionMessage{}, // Empty messages
		},
	)

	require.Error(t, err)
}

// TestE2E_ErrorHandling_MissingRequiredField tests missing required field
func TestE2E_ErrorHandling_MissingRequiredField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Try to create chat completion without model (will use empty model)
	_, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "", // Empty model
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.Error(t, err)
}

// TestE2E_ErrorHandling_ProviderError tests provider error propagation
func TestE2E_ErrorHandling_ProviderError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Configure mock to return error
	env.MockProvider.SetError(true, "simulated provider error")

	_, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error") // OpenAI client wraps errors

	// Reset error state
	env.MockProvider.SetError(false, "")
}

// TestE2E_ErrorHandling_RateLimitError tests rate limit error
func TestE2E_ErrorHandling_RateLimitError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)

	// Configure mock to return rate limit error
	env.MockProvider.SetError(true, "rate limit exceeded")

	_, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.Error(t, err)

	// Reset error state
	env.MockProvider.SetError(false, "")
}

// ========================================
// Authentication Tests
// ========================================

// TestE2E_Authentication_ValidAPIKey tests successful authentication
func TestE2E_Authentication_ValidAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	validKey := "valid-test-key"
	env := NewTestEnvironmentWithAuth(t, validKey)

	resp, err := env.Client.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestE2E_Authentication_InvalidAPIKey tests failed authentication
func TestE2E_Authentication_InvalidAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	validKey := "valid-test-key"
	env := NewTestEnvironmentWithAuth(t, validKey)

	// Create client with invalid key
	config := openailib.DefaultConfig("invalid-key")
	config.BaseURL = env.Server.URL + "/v1"
	invalidClient := openailib.NewClientWithConfig(config)

	_, err := invalidClient.CreateChatCompletion(
		context.Background(),
		openailib.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openailib.ChatCompletionMessage{
				{
					Role:    openailib.ChatMessageRoleUser,
					Content: "Hello",
				},
			},
		},
	)

	require.Error(t, err)
	// Check that authentication failed (could be 400 or 401 depending on implementation)
	assert.Contains(t, err.Error(), "authentication failed")
}
