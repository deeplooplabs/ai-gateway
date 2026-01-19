package hook

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deeplooplabs/ai-gateway/openai"
)

// Hook is the base interface for all hooks
type Hook interface {
	// Name returns the unique name of this hook
	Name() string
}

// AuthenticationHook is called to authenticate API keys
type AuthenticationHook interface {
	Hook
	// Authenticate validates the API key and returns (success, userID, error)
	Authenticate(ctx context.Context, apiKey string) (bool, string, error)
}

// RequestHook is called before/after sending request to provider
type RequestHook interface {
	Hook
	// BeforeRequest is called before sending request (can modify request)
	BeforeRequest(ctx context.Context, req *openai.ChatCompletionRequest) error
	// AfterRequest is called after receiving response (can modify response)
	AfterRequest(ctx context.Context, req *openai.ChatCompletionRequest, resp *openai.ChatCompletionResponse) error
}

// StreamingHook is called for each streaming chunk
type StreamingHook interface {
	Hook
	// OnChunk is called for each SSE chunk in streaming responses
	// Returns the (potentially modified) chunk data
	OnChunk(ctx context.Context, chunk []byte) ([]byte, error)
}

// ErrorHook is called when an error occurs
type ErrorHook interface {
	Hook
	// OnError is called when an error occurs during request processing
	OnError(ctx context.Context, err error)
}

// Registry manages registered hooks
type Registry struct {
	hooks               []Hook
	authenticationHooks []AuthenticationHook
	requestHooks        []RequestHook
	streamingHooks      []StreamingHook
	errorHooks          []ErrorHook
}

// NewRegistry creates a new hook registry
func NewRegistry() *Registry {
	return &Registry{
		hooks:               make([]Hook, 0),
		authenticationHooks: make([]AuthenticationHook, 0),
		requestHooks:        make([]RequestHook, 0),
		streamingHooks:      make([]StreamingHook, 0),
		errorHooks:          make([]ErrorHook, 0),
	}
}

// Register registers a hook based on its concrete type
func (r *Registry) Register(hooks ...Hook) {
	for _, hook := range hooks {
		// Always add to general hooks list
		r.hooks = append(r.hooks, hook)

		// Also add to specific type lists if applicable
		switch h := hook.(type) {
		case AuthenticationHook:
			r.authenticationHooks = append(r.authenticationHooks, h)
		case RequestHook:
			r.requestHooks = append(r.requestHooks, h)
		case StreamingHook:
			r.streamingHooks = append(r.streamingHooks, h)
		case ErrorHook:
			r.errorHooks = append(r.errorHooks, h)
		default:
			slog.Warn(fmt.Sprintf("unknown hook type: %T", h))
		}
	}
}

// AuthenticationHooks returns all authentication hooks
func (r *Registry) AuthenticationHooks() []AuthenticationHook {
	return r.authenticationHooks
}

// RequestHooks returns all request hooks
func (r *Registry) RequestHooks() []RequestHook {
	return r.requestHooks
}

// StreamingHooks returns all streaming hooks
func (r *Registry) StreamingHooks() []StreamingHook {
	return r.streamingHooks
}

// ErrorHooks returns all error hooks
func (r *Registry) ErrorHooks() []ErrorHook {
	return r.errorHooks
}

// All returns all registered hooks
func (r *Registry) All() []Hook {
	return r.hooks
}
