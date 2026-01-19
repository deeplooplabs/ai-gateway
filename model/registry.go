package model

import (
	"github.com/deeplooplabs/ai-gateway/provider"
)

// ProviderRewrite represents a provider and optional model name rewrite
type ProviderRewrite struct {
	Provider     provider.Provider
	ModelRewrite string
}

// ModelRegistry resolves model names to providers
type ModelRegistry interface {
	// Resolve returns the provider and optional model rewrite for a given model name
	Resolve(model string) (provider.Provider, string)
}

// MapModelRegistry is an in-memory model registry
type MapModelRegistry struct {
	models map[string]ProviderRewrite
}

// NewMapModelRegistry creates a new map-based model registry
func NewMapModelRegistry() *MapModelRegistry {
	return &MapModelRegistry{
		models: make(map[string]ProviderRewrite),
	}
}

// Register registers a model with its provider and optional model rewrite
func (r *MapModelRegistry) Register(model string, prov provider.Provider, modelRewrite string) {
	r.models[model] = ProviderRewrite{
		Provider:     prov,
		ModelRewrite: modelRewrite,
	}
}

// Resolve returns the provider and model rewrite for a given model name
func (r *MapModelRegistry) Resolve(model string) (provider.Provider, string) {
	if pr, ok := r.models[model]; ok {
		return pr.Provider, pr.ModelRewrite
	}
	return nil, ""
}
