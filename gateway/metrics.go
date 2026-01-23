package gateway

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the gateway
type Metrics struct {
	RequestsTotal        *prometheus.CounterVec
	RequestDuration      *prometheus.HistogramVec
	TokensUsed           *prometheus.CounterVec
	ErrorsTotal          *prometheus.CounterVec
	ActiveRequests       prometheus.Gauge
	CacheHits            *prometheus.CounterVec
	CacheMisses          *prometheus.CounterVec
	RateLimitExceeded    *prometheus.CounterVec
	ProviderRequestTotal *prometheus.CounterVec
}

// NewMetrics creates a new Metrics instance with Prometheus collectors
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "aigateway"
	}

	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of requests processed",
			},
			[]string{"method", "endpoint", "status", "model"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "Request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"method", "endpoint", "model"},
		),
		TokensUsed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tokens_used_total",
				Help:      "Total number of tokens used",
			},
			[]string{"model", "type"}, // type: input, output, total
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"method", "endpoint", "error_type", "model"},
		),
		ActiveRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_requests",
				Help:      "Number of requests currently being processed",
			},
		),
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"endpoint"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"endpoint"},
		),
		RateLimitExceeded: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "rate_limit_exceeded_total",
				Help:      "Total number of rate limit exceeded errors",
			},
			[]string{"tenant_id", "endpoint"},
		),
		ProviderRequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "provider_requests_total",
				Help:      "Total number of requests sent to providers",
			},
			[]string{"provider", "status"},
		),
	}
}
