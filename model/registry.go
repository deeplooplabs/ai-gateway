package model

import (
	"sync"

	"github.com/deeplooplabs/ai-gateway/provider"
)

// ProviderRewrite represents a provider and optional model name rewrite
type ProviderRewrite struct {
	Provider     provider.Provider
	ModelRewrite string
	PreferredAPI provider.APIType // Optional: preferred API type for this model
}

// ModelRegistry resolves model names to providers
type ModelRegistry interface {
	// Resolve returns the provider and optional model rewrite for a given model name
	Resolve(model string) (provider.Provider, string)
	// ResolveWithAPI returns the provider, model rewrite, and preferred API type
	ResolveWithAPI(model string) (provider.Provider, string, provider.APIType)
}

// MapModelRegistry is an in-memory model registry
type MapModelRegistry struct {
	mu     sync.RWMutex
	models map[string]ProviderRewrite
}

// NewMapModelRegistry creates a new map-based model registry
func NewMapModelRegistry() *MapModelRegistry {
	return &MapModelRegistry{
		models: make(map[string]ProviderRewrite),
	}
}

// RegisterOption is an option for registering a model
type RegisterOption func(*ProviderRewrite)

// WithModelRewrite sets the model rewrite
func WithModelRewrite(rewrite string) RegisterOption {
	return func(pr *ProviderRewrite) {
		pr.ModelRewrite = rewrite
	}
}

// WithPreferredAPI sets the preferred API type
func WithPreferredAPI(apiType provider.APIType) RegisterOption {
	return func(pr *ProviderRewrite) {
		pr.PreferredAPI = apiType
	}
}

// Register registers a model with its provider
func (r *MapModelRegistry) Register(model string, prov provider.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models[model] = ProviderRewrite{
		Provider:     prov,
		ModelRewrite: "",
		PreferredAPI: 0, // Auto-detect from provider
	}
}

// RegisterWithOptions registers a model with its provider and options
func (r *MapModelRegistry) RegisterWithOptions(model string, prov provider.Provider, opts ...RegisterOption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pr := ProviderRewrite{
		Provider:     prov,
		ModelRewrite: "",
		PreferredAPI: 0,
	}
	for _, opt := range opts {
		opt(&pr)
	}
	r.models[model] = pr
}

// Resolve returns the provider and model rewrite for a given model name
func (r *MapModelRegistry) Resolve(model string) (provider.Provider, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if pr, ok := r.models[model]; ok {
		return pr.Provider, pr.ModelRewrite
	}
	return nil, ""
}

// ResolveWithAPI returns the provider, model rewrite, and preferred API type for a given model name
func (r *MapModelRegistry) ResolveWithAPI(model string) (provider.Provider, string, provider.APIType) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if pr, ok := r.models[model]; ok {
		apiType := pr.PreferredAPI
		if apiType == 0 {
			// Auto-detect from provider
			apiType = pr.Provider.SupportedAPIs()
		}
		return pr.Provider, pr.ModelRewrite, apiType
	}
	return nil, "", 0
}
