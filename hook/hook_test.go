package hook

import (
	"context"
	"testing"

	"github.com/deeplooplabs/ai-gateway/provider/openai"
)

// mockHook implements Hook interface for testing
type mockHook struct {
	name string
}

func (m *mockHook) Name() string {
	return m.name
}

func TestHookRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	h1 := &mockHook{name: "hook1"}
	h2 := &mockHook{name: "hook2"}

	registry.Register(h1)
	registry.Register(h2)

	if len(registry.All()) != 2 {
		t.Errorf("expected 2 hooks, got %d", len(registry.All()))
	}
}

// mockAuthHook implements AuthenticationHook
type mockAuthHook struct {
	mockHook
	authenticateFunc func(ctx context.Context, apiKey string) (bool, string, error)
}

func (m *mockAuthHook) Authenticate(ctx context.Context, apiKey string) (bool, string, error) {
	if m.authenticateFunc != nil {
		return m.authenticateFunc(ctx, apiKey)
	}
	return true, "", nil
}

func TestAuthenticationHook(t *testing.T) {
	registry := NewRegistry()

	called := false
	h := &mockAuthHook{
		mockHook: mockHook{name: "auth"},
		authenticateFunc: func(ctx context.Context, apiKey string) (bool, string, error) {
			called = true
			if apiKey == "valid-key" {
				return true, "user-123", nil
			}
			return false, "", nil
		},
	}

	registry.Register(h)

	// Test successful authentication
	success, _, err := h.Authenticate(context.Background(), "valid-key")
	if !success || err != nil {
		t.Error("expected successful authentication")
	}
	if !called {
		t.Error("Authenticate should have been called")
	}

	// Test failed authentication
	success, _, _ = h.Authenticate(context.Background(), "invalid-key")
	if success {
		t.Error("expected failed authentication")
	}
}

// mockRequestHook implements RequestHook
type mockRequestHook struct {
	mockHook
	beforeFunc func(ctx context.Context, req *openai.ChatCompletionRequest) error
	afterFunc  func(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error
}

func (m *mockRequestHook) BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error {
	if m.beforeFunc != nil {
		return m.beforeFunc(ctx, req)
	}
	return nil
}

func (m *mockRequestHook) AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error {
	if m.afterFunc != nil {
		return m.afterFunc(ctx, req, resp)
	}
	return nil
}

func TestRequestHook(t *testing.T) {
	calledBefore := false
	h := &mockRequestHook{
		mockHook: mockHook{name: "request"},
		beforeFunc: func(ctx context.Context, req *openai.ChatCompletionRequest) error {
			calledBefore = true
			req.Model = "modified-model"
			return nil
		},
	}

	req := &openai.ChatCompletionRequest{Model: "gpt-4"}
	h.BeforeRequest(context.Background(), req)

	if !calledBefore {
		t.Error("BeforeRequest should have been called")
	}
	if req.Model != "modified-model" {
		t.Error("BeforeRequest should modify request")
	}
}
