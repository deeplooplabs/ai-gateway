package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	config := &Config{
		RequestsPerSecond: 10,
		Burst:             20,
		Enabled:           true,
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()
	
	// Should allow first request
	if !limiter.Allow(ctx, "user1") {
		t.Fatal("Expected first request to be allowed")
	}
	
	// Should allow burst requests
	for i := 0; i < 19; i++ {
		if !limiter.Allow(ctx, "user1") {
			t.Fatalf("Expected request %d to be allowed", i+2)
		}
	}
	
	// Should deny after burst exhausted
	if limiter.Allow(ctx, "user1") {
		t.Fatal("Expected request to be denied after burst")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	config := &Config{
		RequestsPerSecond: 10,
		Burst:             10,
		Enabled:           true,
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()
	
	// Exhaust tokens
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx, "user1")
	}
	
	// Should be denied
	if limiter.Allow(ctx, "user1") {
		t.Fatal("Expected request to be denied")
	}
	
	// Wait for refill (0.1 second should add 1 token at 10/sec)
	time.Sleep(150 * time.Millisecond)
	
	// Should be allowed after refill
	if !limiter.Allow(ctx, "user1") {
		t.Fatal("Expected request to be allowed after refill")
	}
}

func TestTokenBucket_MultipleKeys(t *testing.T) {
	config := &Config{
		RequestsPerSecond: 10,
		Burst:             5,
		Enabled:           true,
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()
	
	// Exhaust user1
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "user1")
	}
	
	// user1 should be denied
	if limiter.Allow(ctx, "user1") {
		t.Fatal("Expected user1 to be rate limited")
	}
	
	// user2 should still be allowed
	if !limiter.Allow(ctx, "user2") {
		t.Fatal("Expected user2 to be allowed")
	}
}

func TestTokenBucket_Reset(t *testing.T) {
	config := &Config{
		RequestsPerSecond: 10,
		Burst:             5,
		Enabled:           true,
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()
	
	// Exhaust tokens
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "user1")
	}
	
	// Should be denied
	if limiter.Allow(ctx, "user1") {
		t.Fatal("Expected request to be denied")
	}
	
	// Reset
	limiter.Reset(ctx, "user1")
	
	// Should be allowed after reset
	if !limiter.Allow(ctx, "user1") {
		t.Fatal("Expected request to be allowed after reset")
	}
}

func TestTokenBucket_Disabled(t *testing.T) {
	config := &Config{
		RequestsPerSecond: 1,
		Burst:             1,
		Enabled:           false, // Disabled
	}
	limiter := NewTokenBucket(config)
	ctx := context.Background()
	
	// All requests should be allowed when disabled
	for i := 0; i < 100; i++ {
		if !limiter.Allow(ctx, "user1") {
			t.Fatal("Expected all requests to be allowed when disabled")
		}
	}
}
