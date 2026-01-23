package loadbalancer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deeplooplabs/ai-gateway/provider"
)

// Strategy defines the load balancing strategy
type Strategy int

const (
	// RoundRobin distributes requests evenly across providers
	RoundRobin Strategy = iota
	// Random selects a random provider for each request
	Random
	// WeightedRandom selects providers based on their weights
	WeightedRandom
	// LeastConnections selects the provider with fewest active connections
	LeastConnections
)

// ProviderWithWeight wraps a provider with weight and health information
type ProviderWithWeight struct {
	Provider         provider.Provider
	Weight           int           // Weight for weighted random (default: 1)
	ActiveRequests   int32         // Current active requests
	TotalRequests    uint64        // Total requests handled
	TotalErrors      uint64        // Total errors encountered
	LastHealthCheck  time.Time     // Last health check time
	Healthy          bool          // Health status
	HealthCheckURL   string        // Optional health check endpoint
	HealthCheckInterval time.Duration // Health check interval (default: 30s)
}

// LoadBalancedProvider wraps multiple providers with load balancing
type LoadBalancedProvider struct {
	name      string
	providers []*ProviderWithWeight
	strategy  Strategy
	counter   uint64 // For round-robin
	mu        sync.RWMutex
	
	// Health check configuration
	healthCheckEnabled  bool
	healthCheckInterval time.Duration
	stopHealthCheck     chan struct{}
}

// Config holds load balancer configuration
type Config struct {
	Name                string
	Strategy            Strategy
	Providers           []provider.Provider
	Weights             []int  // Optional weights for WeightedRandom
	HealthCheckEnabled  bool
	HealthCheckInterval time.Duration
}

// DefaultConfig returns a default load balancer configuration
func DefaultConfig(name string) *Config {
	return &Config{
		Name:                name,
		Strategy:            RoundRobin,
		HealthCheckEnabled:  false,
		HealthCheckInterval: 30 * time.Second,
	}
}

// New creates a new load-balanced provider
func New(config *Config) (*LoadBalancedProvider, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	
	if len(config.Providers) == 0 {
		return nil, errors.New("at least one provider is required")
	}
	
	// Create provider wrappers
	providerWrappers := make([]*ProviderWithWeight, len(config.Providers))
	for i, p := range config.Providers {
		weight := 1
		if len(config.Weights) > i {
			weight = config.Weights[i]
		}
		
		providerWrappers[i] = &ProviderWithWeight{
			Provider:            p,
			Weight:              weight,
			Healthy:             true,
			HealthCheckInterval: config.HealthCheckInterval,
			LastHealthCheck:     time.Now(),
		}
	}
	
	lb := &LoadBalancedProvider{
		name:                config.Name,
		providers:           providerWrappers,
		strategy:            config.Strategy,
		healthCheckEnabled:  config.HealthCheckEnabled,
		healthCheckInterval: config.HealthCheckInterval,
		stopHealthCheck:     make(chan struct{}),
	}
	
	// Start health checks if enabled
	if lb.healthCheckEnabled {
		go lb.runHealthChecks()
	}
	
	return lb, nil
}

// Name returns the provider name
func (lb *LoadBalancedProvider) Name() string {
	return lb.name
}

// SupportedAPIs returns the supported APIs (union of all providers)
func (lb *LoadBalancedProvider) SupportedAPIs() provider.APIType {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	if len(lb.providers) == 0 {
		return 0
	}
	
	// Return the API type of the first provider
	// (assumes all providers support the same APIs)
	return lb.providers[0].Provider.SupportedAPIs()
}

// SendRequest sends a request using the load balancing strategy
func (lb *LoadBalancedProvider) SendRequest(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	p, err := lb.selectProvider()
	if err != nil {
		return nil, err
	}
	
	// Track active requests
	atomic.AddInt32(&p.ActiveRequests, 1)
	atomic.AddUint64(&p.TotalRequests, 1)
	defer atomic.AddInt32(&p.ActiveRequests, -1)
	
	// Send request
	resp, err := p.Provider.SendRequest(ctx, req)
	if err != nil {
		atomic.AddUint64(&p.TotalErrors, 1)
		// Mark as unhealthy if too many errors
		if atomic.LoadUint64(&p.TotalErrors) > 10 {
			lb.mu.Lock()
			p.Healthy = false
			lb.mu.Unlock()
		}
		return nil, err
	}
	
	return resp, nil
}

// selectProvider selects a provider based on the load balancing strategy
func (lb *LoadBalancedProvider) selectProvider() (*ProviderWithWeight, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	// Filter healthy providers
	healthyProviders := make([]*ProviderWithWeight, 0, len(lb.providers))
	for _, p := range lb.providers {
		if p.Healthy {
			healthyProviders = append(healthyProviders, p)
		}
	}
	
	if len(healthyProviders) == 0 {
		return nil, errors.New("no healthy providers available")
	}
	
	switch lb.strategy {
	case RoundRobin:
		return lb.selectRoundRobin(healthyProviders), nil
	case Random:
		return lb.selectRandom(healthyProviders), nil
	case WeightedRandom:
		return lb.selectWeightedRandom(healthyProviders), nil
	case LeastConnections:
		return lb.selectLeastConnections(healthyProviders), nil
	default:
		return lb.selectRoundRobin(healthyProviders), nil
	}
}

// selectRoundRobin selects provider using round-robin
func (lb *LoadBalancedProvider) selectRoundRobin(providers []*ProviderWithWeight) *ProviderWithWeight {
	count := atomic.AddUint64(&lb.counter, 1)
	return providers[int(count-1)%len(providers)]
}

// selectRandom selects a random provider
func (lb *LoadBalancedProvider) selectRandom(providers []*ProviderWithWeight) *ProviderWithWeight {
	return providers[time.Now().UnixNano()%int64(len(providers))]
}

// selectWeightedRandom selects provider based on weights
func (lb *LoadBalancedProvider) selectWeightedRandom(providers []*ProviderWithWeight) *ProviderWithWeight {
	// Calculate total weight
	totalWeight := 0
	for _, p := range providers {
		totalWeight += p.Weight
	}
	
	// Select random value
	random := int(time.Now().UnixNano() % int64(totalWeight))
	
	// Find provider based on weight
	sum := 0
	for _, p := range providers {
		sum += p.Weight
		if random < sum {
			return p
		}
	}
	
	return providers[0]
}

// selectLeastConnections selects provider with fewest active connections
func (lb *LoadBalancedProvider) selectLeastConnections(providers []*ProviderWithWeight) *ProviderWithWeight {
	minConnections := atomic.LoadInt32(&providers[0].ActiveRequests)
	selected := providers[0]
	
	for i := 1; i < len(providers); i++ {
		connections := atomic.LoadInt32(&providers[i].ActiveRequests)
		if connections < minConnections {
			minConnections = connections
			selected = providers[i]
		}
	}
	
	return selected
}

// runHealthChecks periodically checks provider health
func (lb *LoadBalancedProvider) runHealthChecks() {
	ticker := time.NewTicker(lb.healthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lb.checkHealth()
		case <-lb.stopHealthCheck:
			return
		}
	}
}

// checkHealth checks health of all providers
func (lb *LoadBalancedProvider) checkHealth() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	for _, p := range lb.providers {
		// Simple health check: if error rate is low, mark as healthy
		totalRequests := atomic.LoadUint64(&p.TotalRequests)
		totalErrors := atomic.LoadUint64(&p.TotalErrors)
		
		if totalRequests > 0 {
			errorRate := float64(totalErrors) / float64(totalRequests)
			p.Healthy = errorRate < 0.5 // Mark healthy if error rate < 50%
		}
		
		p.LastHealthCheck = time.Now()
	}
}

// Close stops the health check goroutine
func (lb *LoadBalancedProvider) Close() error {
	if lb.healthCheckEnabled {
		close(lb.stopHealthCheck)
	}
	return nil
}

// GetStats returns statistics for all providers
func (lb *LoadBalancedProvider) GetStats() []ProviderStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	stats := make([]ProviderStats, len(lb.providers))
	for i, p := range lb.providers {
		stats[i] = ProviderStats{
			Name:           p.Provider.Name(),
			Weight:         p.Weight,
			Healthy:        p.Healthy,
			ActiveRequests: atomic.LoadInt32(&p.ActiveRequests),
			TotalRequests:  atomic.LoadUint64(&p.TotalRequests),
			TotalErrors:    atomic.LoadUint64(&p.TotalErrors),
			LastHealthCheck: p.LastHealthCheck,
		}
	}
	
	return stats
}

// ProviderStats represents statistics for a single provider
type ProviderStats struct {
	Name            string
	Weight          int
	Healthy         bool
	ActiveRequests  int32
	TotalRequests   uint64
	TotalErrors     uint64
	LastHealthCheck time.Time
}
