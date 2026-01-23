package cache

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// lruCache implements an LRU cache with TTL support
type lruCache struct {
	mu       sync.RWMutex
	config   *Config
	items    map[string]*list.Element
	lruList  *list.List
	size     int64
	hits     uint64
	misses   uint64
}

// cacheEntry represents a single cache entry
type cacheEntry struct {
	key       string
	value     []byte
	expiresAt time.Time
	size      int64
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(config *Config) Cache {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &lruCache{
		config:  config,
		items:   make(map[string]*list.Element),
		lruList: list.New(),
	}
}

// Get retrieves a value from the cache
func (c *lruCache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	element, found := c.items[key]
	c.mu.RUnlock()
	
	if !found {
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}
	
	entry := element.Value.(*cacheEntry)
	
	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		c.removeElement(element)
		c.mu.Unlock()
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}
	
	// Move to front (most recently used)
	c.mu.Lock()
	c.lruList.MoveToFront(element)
	c.mu.Unlock()
	
	atomic.AddUint64(&c.hits, 1)
	return entry.value, true
}

// Set stores a value in the cache
func (c *lruCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.config.Enabled {
		return nil
	}
	
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}
	
	entrySize := int64(len(key) + len(value))
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if key already exists
	if element, found := c.items[key]; found {
		// Update existing entry
		entry := element.Value.(*cacheEntry)
		c.size -= entry.size
		entry.value = value
		entry.expiresAt = time.Now().Add(ttl)
		entry.size = entrySize
		c.size += entrySize
		c.lruList.MoveToFront(element)
		return nil
	}
	
	// Evict if necessary
	for c.size+entrySize > c.config.MaxSize || c.lruList.Len() >= c.config.MaxItems {
		if c.lruList.Len() == 0 {
			break
		}
		c.removeElement(c.lruList.Back())
	}
	
	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
		size:      entrySize,
	}
	
	element := c.lruList.PushFront(entry)
	c.items[key] = element
	c.size += entrySize
	
	return nil
}

// Delete removes a value from the cache
func (c *lruCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if element, found := c.items[key]; found {
		c.removeElement(element)
	}
	
	return nil
}

// Clear removes all values from the cache
func (c *lruCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*list.Element)
	c.lruList = list.New()
	c.size = 0
	
	return nil
}

// Stats returns cache statistics
func (c *lruCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return CacheStats{
		Hits:   atomic.LoadUint64(&c.hits),
		Misses: atomic.LoadUint64(&c.misses),
		Size:   uint64(c.size),
		Items:  uint64(c.lruList.Len()),
	}
}

// removeElement removes an element from the cache (must be called with lock held)
func (c *lruCache) removeElement(element *list.Element) {
	entry := element.Value.(*cacheEntry)
	delete(c.items, entry.key)
	c.lruList.Remove(element)
	c.size -= entry.size
}
