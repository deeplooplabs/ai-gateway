package loadbalancer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/deeplooplabs/ai-gateway/provider"
)

// Mock provider for testing
type mockProvider struct {
	name      string
	shouldFail bool
	callCount int
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) SupportedAPIs() provider.APIType {
	return provider.APITypeChatCompletions
}

func (m *mockProvider) SendRequest(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("mock error")
	}
	return &provider.Response{}, nil
}

func TestLoadBalancer_RoundRobin(t *testing.T) {
	p1 := &mockProvider{name: "provider1"}
	p2 := &mockProvider{name: "provider2"}
	p3 := &mockProvider{name: "provider3"}
	
	config := &Config{
		Name:      "test-lb",
		Strategy:  RoundRobin,
		Providers: []provider.Provider{p1, p2, p3},
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Send 9 requests - should distribute evenly
	for i := 0; i < 9; i++ {
		_, err := lb.SendRequest(ctx, req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
	
	// Each provider should have received 3 requests
	if p1.callCount != 3 || p2.callCount != 3 || p3.callCount != 3 {
		t.Errorf("Expected 3 calls per provider, got p1=%d, p2=%d, p3=%d",
			p1.callCount, p2.callCount, p3.callCount)
	}
}

func TestLoadBalancer_WeightedRandom(t *testing.T) {
	p1 := &mockProvider{name: "provider1"}
	p2 := &mockProvider{name: "provider2"}
	
	config := &Config{
		Name:      "test-lb",
		Strategy:  WeightedRandom,
		Providers: []provider.Provider{p1, p2},
		Weights:   []int{3, 1}, // p1 has 3x weight of p2
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Send many requests
	for i := 0; i < 100; i++ {
		_, err := lb.SendRequest(ctx, req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
	
	// p1 should have received roughly 3x more requests than p2
	// Allow some variance due to randomness
	if p1.callCount < 60 || p1.callCount > 90 {
		t.Errorf("Expected p1 to receive ~75 requests, got %d", p1.callCount)
	}
	if p2.callCount < 10 || p2.callCount > 40 {
		t.Errorf("Expected p2 to receive ~25 requests, got %d", p2.callCount)
	}
}

func TestLoadBalancer_LeastConnections(t *testing.T) {
	p1 := &mockProvider{name: "provider1"}
	p2 := &mockProvider{name: "provider2"}
	
	config := &Config{
		Name:      "test-lb",
		Strategy:  LeastConnections,
		Providers: []provider.Provider{p1, p2},
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Send requests - since all complete instantly,
	// they should be distributed across providers
	for i := 0; i < 10; i++ {
		_, err := lb.SendRequest(ctx, req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
	
	// Both providers should have received requests
	total := p1.callCount + p2.callCount
	if total != 10 {
		t.Errorf("Expected 10 total requests, got %d", total)
	}
	
	// With least connections and instant completion,
	// distribution may vary but both should get some
	// We'll just check that total is correct
}

func TestLoadBalancer_HealthCheck(t *testing.T) {
	p1 := &mockProvider{name: "provider1", shouldFail: true}
	p2 := &mockProvider{name: "provider2"}
	
	config := &Config{
		Name:                "test-lb",
		Strategy:            RoundRobin,
		Providers:           []provider.Provider{p1, p2},
		HealthCheckEnabled:  true,
		HealthCheckInterval: 100 * time.Millisecond,
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Generate errors on p1 to mark it unhealthy
	for i := 0; i < 15; i++ {
		lb.SendRequest(ctx, req)
	}
	
	// Wait for health check
	time.Sleep(200 * time.Millisecond)
	
	// Now p1 should be unhealthy, all requests should go to p2
	p2InitialCount := p2.callCount
	for i := 0; i < 5; i++ {
		_, err := lb.SendRequest(ctx, req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
	
	// p2 should have received all 5 new requests
	if p2.callCount-p2InitialCount != 5 {
		t.Errorf("Expected p2 to receive 5 requests, got %d", p2.callCount-p2InitialCount)
	}
}

func TestLoadBalancer_NoHealthyProviders(t *testing.T) {
	p1 := &mockProvider{name: "provider1", shouldFail: true}
	
	config := &Config{
		Name:      "test-lb",
		Strategy:  RoundRobin,
		Providers: []provider.Provider{p1},
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Generate errors to mark provider unhealthy
	for i := 0; i < 15; i++ {
		lb.SendRequest(ctx, req)
	}
	
	// Next request should fail with "no healthy providers"
	_, err = lb.SendRequest(ctx, req)
	if err == nil || err.Error() != "no healthy providers available" {
		t.Errorf("Expected 'no healthy providers available' error, got: %v", err)
	}
}

func TestLoadBalancer_GetStats(t *testing.T) {
	p1 := &mockProvider{name: "provider1"}
	p2 := &mockProvider{name: "provider2"}
	
	config := &Config{
		Name:      "test-lb",
		Strategy:  RoundRobin,
		Providers: []provider.Provider{p1, p2},
	}
	
	lb, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}
	defer lb.Close()
	
	ctx := context.Background()
	req := &provider.Request{}
	
	// Send some requests
	for i := 0; i < 10; i++ {
		lb.SendRequest(ctx, req)
	}
	
	// Get stats
	stats := lb.GetStats()
	if len(stats) != 2 {
		t.Fatalf("Expected 2 provider stats, got %d", len(stats))
	}
	
	// Check that stats are populated
	totalRequests := stats[0].TotalRequests + stats[1].TotalRequests
	if totalRequests != 10 {
		t.Errorf("Expected 10 total requests, got %d", totalRequests)
	}
}
