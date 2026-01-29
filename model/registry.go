package model

import (
	"fmt"
	"log/slog"
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
	// ListModels returns a list of all registered model names
	ListModels() []string
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
	providerName := "unknown"
	if prov != nil {
		providerName = prov.Name()
	}
	r.models[model] = ProviderRewrite{
		Provider:     prov,
		ModelRewrite: "",
		PreferredAPI: 0, // Auto-detect from provider
	}
	logModelRegistration(model, "", providerName, prov.SupportedAPIs())
}

// RegisterWithOptions registers a model with its provider and options
func (r *MapModelRegistry) RegisterWithOptions(model string, prov provider.Provider, opts ...RegisterOption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	providerName := "unknown"
	if prov != nil {
		providerName = prov.Name()
	}
	pr := ProviderRewrite{
		Provider:     prov,
		ModelRewrite: "",
		PreferredAPI: 0,
	}
	for _, opt := range opts {
		opt(&pr)
	}
	r.models[model] = pr
	logModelRegistration(model, pr.ModelRewrite, providerName, pr.PreferredAPI)
}

// logModelRegistration logs the model registration with type information
func logModelRegistration(model, modelRewrite, providerName string, apiType provider.APIType) {
	apiTypeStr := formatAPIType(apiType)

	msg := "Registered model"
	attrs := []any{
		"model", model,
		"type", apiTypeStr,
		"provider", providerName,
	}

	// Add model rewrite if present
	if modelRewrite != "" {
		attrs = append(attrs, "rewrite_to", modelRewrite)
	}

	slog.Info(msg, attrs...)
}

// formatAPIType formats the API type into a human-readable string
func formatAPIType(apiType provider.APIType) string {
	switch {
	case apiType == 0:
		return "auto"
	case apiType.Supports(provider.APITypeChatCompletions) && apiType.Supports(provider.APITypeEmbeddings):
		return "chat+embeddings"
	case apiType.Supports(provider.APITypeChatCompletions) && apiType.Supports(provider.APITypeResponses):
		return "chat+responses"
	case apiType.Supports(provider.APITypeChatCompletions):
		return "chat"
	case apiType.Supports(provider.APITypeEmbeddings):
		return "embedding"
	case apiType.Supports(provider.APITypeImages):
		return "image"
	case apiType.Supports(provider.APITypeResponses):
		return "response"
	default:
		return fmt.Sprintf("unknown(%d)", apiType)
	}
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

// ListModels returns a list of all registered model names
func (r *MapModelRegistry) ListModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	models := make([]string, 0, len(r.models))
	for model := range r.models {
		models = append(models, model)
	}
	return models
}
