package cache

import (
	"context"
	"testing"
	"time"
)

func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(DefaultConfig())
	ctx := context.Background()
	
	// Set a value
	key := "test-key"
	value := []byte("test-value")
	err := cache.Set(ctx, key, value, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	
	// Get the value
	retrieved, found := cache.Get(ctx, key)
	if !found {
		t.Fatal("Expected to find key in cache")
	}
	
	if string(retrieved) != string(value) {
		t.Fatalf("Expected %s, got %s", value, retrieved)
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(DefaultConfig())
	ctx := context.Background()
	
	// Set a value with short TTL
	key := "expire-key"
	value := []byte("expire-value")
	err := cache.Set(ctx, key, value, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	
	// Should be found immediately
	_, found := cache.Get(ctx, key)
	if !found {
		t.Fatal("Expected to find key in cache")
	}
	
	// Wait for expiration
	time.Sleep(200 * time.Millisecond)
	
	// Should not be found
	_, found = cache.Get(ctx, key)
	if found {
		t.Fatal("Expected key to be expired")
	}
}

func TestLRUCache_LRUEviction(t *testing.T) {
	config := &Config{
		MaxSize:    1024,
		MaxItems:   2, // Only allow 2 items
		DefaultTTL: 5 * time.Minute,
		Enabled:    true,
	}
	cache := NewLRUCache(config)
	ctx := context.Background()
	
	// Add 3 items (should evict the first one)
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	cache.Set(ctx, "key2", []byte("value2"), 5*time.Minute)
	cache.Set(ctx, "key3", []byte("value3"), 5*time.Minute)
	
	// key1 should be evicted
	_, found := cache.Get(ctx, "key1")
	if found {
		t.Fatal("Expected key1 to be evicted")
	}
	
	// key2 and key3 should exist
	_, found = cache.Get(ctx, "key2")
	if !found {
		t.Fatal("Expected key2 to exist")
	}
	
	_, found = cache.Get(ctx, "key3")
	if !found {
		t.Fatal("Expected key3 to exist")
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(DefaultConfig())
	ctx := context.Background()
	
	// Initial stats
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Fatal("Expected zero stats initially")
	}
	
	// Set and get
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	cache.Get(ctx, "key1") // hit
	cache.Get(ctx, "key2") // miss
	
	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Fatalf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Fatalf("Expected 1 miss, got %d", stats.Misses)
	}
}
