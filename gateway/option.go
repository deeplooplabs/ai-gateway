package gateway

import (
	"github.com/deeplooplabs/ai-gateway/hook"
	"github.com/deeplooplabs/ai-gateway/model"
)

// Option configures the Gateway
type Option func(*Gateway)

// WithModelRegistry sets the model registry
func WithModelRegistry(registry model.ModelRegistry) Option {
	return func(g *Gateway) {
		g.modelRegistry = registry
	}
}

// WithHooks sets the hook registry
func WithHooks(hooks *hook.Registry) Option {
	return func(g *Gateway) {
		g.hooks = hooks
	}
}

// WithHook registers a single hook
func WithHook(h hook.Hook) Option {
	return func(g *Gateway) {
		g.hooks.Register(h)
	}
}
