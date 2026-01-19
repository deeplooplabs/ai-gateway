package ai_gateway

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Provider represents an AI provider
type Provider interface{}

// Context represents the request context throughout its lifecycle
type Context struct {
	RequestID   string
	StartTime   time.Time
	OriginalReq *http.Request
	Metadata    map[string]any
	Provider    Provider
	mu          sync.RWMutex
}

// NewContext creates a new request context
func NewContext(req *http.Request) *Context {
	return &Context{
		RequestID:   uuid.New().String(),
		StartTime:   time.Now(),
		OriginalReq: req,
		Metadata:    make(map[string]any),
	}
}

// Set stores a value in the context metadata
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Metadata[key] = value
}

// Get retrieves a value from the context metadata
func (c *Context) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Metadata[key]
}
