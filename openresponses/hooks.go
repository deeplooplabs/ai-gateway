package openresponses

import (
	"context"

	"github.com/deeplooplabs/ai-gateway/hook"
)

// RequestHook is called before/after sending request to provider (OpenResponses-specific)
type RequestHook interface {
	hook.Hook
	// BeforeRequest is called before sending request (can modify request)
	BeforeRequest(ctx context.Context, req *CreateRequest) error
	// AfterRequest is called after receiving response (can modify response)
	AfterRequest(ctx context.Context, req *CreateRequest, resp *Response) error
}

// StreamingHook is called for each streaming event (OpenResponses-specific)
type StreamingHook interface {
	hook.Hook
	// OnEvent is called for each streaming event
	// Returns the (potentially modified) event
	OnEvent(ctx context.Context, event StreamingEvent) (StreamingEvent, error)
}

// Registry manages OpenResponses-specific hooks
type Registry struct {
	hooks           []hook.Hook
	requestHooks    []RequestHook
	streamingHooks  []StreamingHook
	parentRegistry  *hook.Registry
}

// NewRegistry creates a new OpenResponses hook registry
func NewRegistry(parent *hook.Registry) *Registry {
	return &Registry{
		hooks:          make([]hook.Hook, 0),
		requestHooks:   make([]RequestHook, 0),
		streamingHooks: make([]StreamingHook, 0),
		parentRegistry: parent,
	}
}

// Register registers OpenResponses-specific hooks
func (r *Registry) Register(hooks ...hook.Hook) {
	for _, h := range hooks {
		// Always add to general hooks list
		r.hooks = append(r.hooks, h)

		// Also add to specific type lists if applicable
		switch hh := h.(type) {
		case RequestHook:
			r.requestHooks = append(r.requestHooks, hh)
		case StreamingHook:
			r.streamingHooks = append(r.streamingHooks, hh)
		default:
			// Try to register with parent registry for backward compatibility
			if r.parentRegistry != nil {
				r.parentRegistry.Register(h)
			}
		}
	}
}

// RequestHooks returns all request hooks
func (r *Registry) RequestHooks() []RequestHook {
	return r.requestHooks
}

// StreamingHooks returns all streaming hooks
func (r *Registry) StreamingHooks() []StreamingHook {
	return r.streamingHooks
}

// All returns all registered hooks
func (r *Registry) All() []hook.Hook {
	return r.hooks
}

// CombineWithParent combines OpenResponses hooks with parent registry hooks
func (r *Registry) CombineWithParent() *hook.Registry {
	combined := hook.NewRegistry()

	// Register all hooks from parent
	if r.parentRegistry != nil {
		for _, h := range r.parentRegistry.All() {
			combined.Register(h)
		}
	}

	// Register all OpenResponses-specific hooks
	for _, h := range r.hooks {
		combined.Register(h)
	}

	return combined
}
