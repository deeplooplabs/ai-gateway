package aigateway

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
	TraceID     string // Distributed tracing trace ID
	SpanID      string // Distributed tracing span ID
	StartTime   time.Time
	OriginalReq *http.Request
	Metadata    map[string]any
	Provider    Provider
	mu          sync.RWMutex
}

// NewContext creates a new request context
func NewContext(req *http.Request) *Context {
	requestID := uuid.New().String()
	
	// Try to extract trace ID from headers (W3C Trace Context or X-Trace-Id)
	traceID := req.Header.Get("traceparent")
	if traceID == "" {
		traceID = req.Header.Get("X-Trace-Id")
	}
	if traceID == "" {
		traceID = requestID // Use request ID as trace ID if not provided
	}
	
	// Extract or generate span ID
	spanID := req.Header.Get("X-Span-Id")
	if spanID == "" {
		spanID = uuid.New().String()
	}
	
	return &Context{
		RequestID:   requestID,
		TraceID:     traceID,
		SpanID:      spanID,
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
