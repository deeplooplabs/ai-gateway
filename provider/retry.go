package provider

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int
	
	// InitialBackoff is the initial backoff duration (default: 100ms)
	InitialBackoff time.Duration
	
	// MaxBackoff is the maximum backoff duration (default: 10s)
	MaxBackoff time.Duration
	
	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
	
	// Jitter adds randomness to backoff (default: true)
	Jitter bool
	
	// RetryableStatusCodes are HTTP status codes that trigger retries
	RetryableStatusCodes map[int]bool
	
	// Enabled indicates whether retries are enabled
	Enabled bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:         3,
		InitialBackoff:     100 * time.Millisecond,
		MaxBackoff:         10 * time.Second,
		BackoffMultiplier:  2.0,
		Jitter:             true,
		RetryableStatusCodes: map[int]bool{
			http.StatusRequestTimeout:      true, // 408
			http.StatusTooManyRequests:     true, // 429
			http.StatusInternalServerError: true, // 500
			http.StatusBadGateway:          true, // 502
			http.StatusServiceUnavailable:  true, // 503
			http.StatusGatewayTimeout:      true, // 504
		},
		Enabled: true,
	}
}

// shouldRetry determines if a request should be retried based on status code
func (rc *RetryConfig) shouldRetry(statusCode int) bool {
	if !rc.Enabled {
		return false
	}
	return rc.RetryableStatusCodes[statusCode]
}

// getBackoffDuration calculates the backoff duration for the given attempt
func (rc *RetryConfig) getBackoffDuration(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(rc.InitialBackoff) * math.Pow(rc.BackoffMultiplier, float64(attempt))
	
	// Cap at max backoff
	if backoff > float64(rc.MaxBackoff) {
		backoff = float64(rc.MaxBackoff)
	}
	
	// Add jitter if enabled
	if rc.Jitter {
		// Add up to 25% random jitter
		jitter := backoff * 0.25 * rand.Float64()
		backoff += jitter
	}
	
	return time.Duration(backoff)
}

// retryWithBackoff executes a function with retry logic
func retryWithBackoff(ctx context.Context, config *RetryConfig, fn func() (*http.Response, error)) (*http.Response, error) {
	if config == nil || !config.Enabled {
		return fn()
	}
	
	var lastErr error
	var resp *http.Response
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		resp, lastErr = fn()
		
		// Check if we should retry
		if lastErr == nil && resp != nil {
			// Check status code
			if !config.shouldRetry(resp.StatusCode) {
				return resp, nil
			}
			// Close response body before retry
			if resp.Body != nil {
				resp.Body.Close()
			}
		}
		
		// Don't sleep after the last attempt
		if attempt < config.MaxRetries {
			// Check context cancellation before sleeping
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(config.getBackoffDuration(attempt)):
				// Continue to next retry
			}
		}
	}
	
	// Return the last response/error
	if lastErr != nil {
		return nil, lastErr
	}
	return resp, nil
}
