package aigateway

import (
	"net/http"
	"testing"
)

func TestNewContext(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
	ctx := NewContext(req)

	if ctx.RequestID == "" {
		t.Error("RequestID should not be empty")
	}
	if ctx.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
	if ctx.OriginalReq != req {
		t.Error("OriginalReq should match")
	}
	if ctx.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestContextSetGet(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Set("key", "value")

	if val := ctx.Get("key"); val != "value" {
		t.Errorf("expected 'value', got '%v'", val)
	}
	if val := ctx.Get("nonexistent"); val != nil {
		t.Errorf("expected nil, got '%v'", val)
	}
}
