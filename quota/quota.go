package quota

import (
	"context"
	"sync"
	"time"
)

// ResetPeriod defines when quotas reset
type ResetPeriod int

const (
	// Hourly resets quotas every hour
	Hourly ResetPeriod = iota
	// Daily resets quotas every day
	Daily
	// Weekly resets quotas every week
	Weekly
	// Monthly resets quotas every month
	Monthly
	// Never means quotas never reset automatically
	Never
)

// Manager tracks and enforces token usage quotas
type Manager interface {
	// RecordUsage records token usage for a tenant
	RecordUsage(ctx context.Context, tenantID string, inputTokens, outputTokens, totalTokens int) error
	
	// CheckQuota checks if tenant has remaining quota
	CheckQuota(ctx context.Context, tenantID string) (bool, *Usage, error)
	
	// GetUsage returns current usage for a tenant
	GetUsage(ctx context.Context, tenantID string) (*Usage, error)
	
	// SetQuota sets the quota limit for a tenant
	SetQuota(ctx context.Context, tenantID string, limit int64) error
	
	// ResetUsage resets usage for a tenant
	ResetUsage(ctx context.Context, tenantID string) error
	
	// ResetAll resets usage for all tenants
	ResetAll(ctx context.Context) error
}

// Usage represents token usage for a tenant
type Usage struct {
	TenantID      string
	InputTokens   int64
	OutputTokens  int64
	TotalTokens   int64
	QuotaLimit    int64 // 0 means unlimited
	ResetAt       time.Time
	LastUpdated   time.Time
}

// Config holds quota manager configuration
type Config struct {
	// DefaultQuota is the default quota for new tenants (0 = unlimited)
	DefaultQuota int64
	
	// ResetPeriod determines when quotas reset
	ResetPeriod ResetPeriod
	
	// Enabled indicates whether quota management is enabled
	Enabled bool
}

// DefaultConfig returns a default quota configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultQuota: 0, // Unlimited by default
		ResetPeriod:  Daily,
		Enabled:      false,
	}
}

// memoryQuotaManager implements in-memory quota management
type memoryQuotaManager struct {
	mu           sync.RWMutex
	config       *Config
	usages       map[string]*Usage
	stopReset    chan struct{}
}

// NewMemoryManager creates a new in-memory quota manager
func NewMemoryManager(config *Config) Manager {
	if config == nil {
		config = DefaultConfig()
	}
	
	mgr := &memoryQuotaManager{
		config:    config,
		usages:    make(map[string]*Usage),
		stopReset: make(chan struct{}),
	}
	
	// Start automatic reset goroutine if period is set
	if config.ResetPeriod != Never {
		go mgr.runAutoReset()
	}
	
	return mgr
}

// RecordUsage records token usage for a tenant
func (m *memoryQuotaManager) RecordUsage(ctx context.Context, tenantID string, inputTokens, outputTokens, totalTokens int) error {
	if !m.config.Enabled {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	usage, exists := m.usages[tenantID]
	if !exists {
		usage = &Usage{
			TenantID:   tenantID,
			QuotaLimit: m.config.DefaultQuota,
			ResetAt:    m.calculateResetTime(),
		}
		m.usages[tenantID] = usage
	}
	
	// Check if reset is needed
	if time.Now().After(usage.ResetAt) {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		usage.TotalTokens = 0
		usage.ResetAt = m.calculateResetTime()
	}
	
	// Record usage
	usage.InputTokens += int64(inputTokens)
	usage.OutputTokens += int64(outputTokens)
	usage.TotalTokens += int64(totalTokens)
	usage.LastUpdated = time.Now()
	
	return nil
}

// CheckQuota checks if tenant has remaining quota
func (m *memoryQuotaManager) CheckQuota(ctx context.Context, tenantID string) (bool, *Usage, error) {
	if !m.config.Enabled {
		return true, nil, nil
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	usage, exists := m.usages[tenantID]
	if !exists {
		// New tenant with default quota
		return true, &Usage{
			TenantID:   tenantID,
			QuotaLimit: m.config.DefaultQuota,
		}, nil
	}
	
	// Check if reset is needed (skip if reset time is zero)
	if !usage.ResetAt.IsZero() && time.Now().After(usage.ResetAt) {
		return true, usage, nil
	}
	
	// Check quota (0 = unlimited)
	if usage.QuotaLimit == 0 {
		return true, usage, nil
	}
	
	hasQuota := usage.TotalTokens < usage.QuotaLimit
	return hasQuota, usage, nil
}

// GetUsage returns current usage for a tenant
func (m *memoryQuotaManager) GetUsage(ctx context.Context, tenantID string) (*Usage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	usage, exists := m.usages[tenantID]
	if !exists {
		return &Usage{
			TenantID:   tenantID,
			QuotaLimit: m.config.DefaultQuota,
			ResetAt:    m.calculateResetTime(),
		}, nil
	}
	
	// Return a copy
	usageCopy := *usage
	return &usageCopy, nil
}

// SetQuota sets the quota limit for a tenant
func (m *memoryQuotaManager) SetQuota(ctx context.Context, tenantID string, limit int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	usage, exists := m.usages[tenantID]
	if !exists {
		usage = &Usage{
			TenantID:   tenantID,
			QuotaLimit: limit,
			ResetAt:    m.calculateResetTime(),
		}
		m.usages[tenantID] = usage
	} else {
		usage.QuotaLimit = limit
	}
	
	return nil
}

// ResetUsage resets usage for a tenant
func (m *memoryQuotaManager) ResetUsage(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	usage, exists := m.usages[tenantID]
	if exists {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		usage.TotalTokens = 0
		usage.ResetAt = m.calculateResetTime()
		usage.LastUpdated = time.Now()
	}
	
	return nil
}

// ResetAll resets usage for all tenants
func (m *memoryQuotaManager) ResetAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	resetAt := m.calculateResetTime()
	now := time.Now()
	
	for _, usage := range m.usages {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		usage.TotalTokens = 0
		usage.ResetAt = resetAt
		usage.LastUpdated = now
	}
	
	return nil
}

// calculateResetTime calculates the next reset time based on the reset period
func (m *memoryQuotaManager) calculateResetTime() time.Time {
	now := time.Now()
	
	switch m.config.ResetPeriod {
	case Hourly:
		return now.Add(1 * time.Hour).Truncate(time.Hour)
	case Daily:
		return now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	case Monthly:
		// Reset on the first day of next month
		year, month, _ := now.Date()
		return time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	case Never:
		return time.Time{} // Zero time means never reset
	default:
		return now.Add(24 * time.Hour)
	}
}

// runAutoReset runs periodic resets
func (m *memoryQuotaManager) runAutoReset() {
	var interval time.Duration
	
	switch m.config.ResetPeriod {
	case Hourly:
		interval = 1 * time.Hour
	case Daily:
		interval = 24 * time.Hour
	default:
		interval = 24 * time.Hour
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.checkAndResetExpired()
		case <-m.stopReset:
			return
		}
	}
}

// checkAndResetExpired checks and resets expired quotas
func (m *memoryQuotaManager) checkAndResetExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	resetAt := m.calculateResetTime()
	
	for _, usage := range m.usages {
		if now.After(usage.ResetAt) {
			usage.InputTokens = 0
			usage.OutputTokens = 0
			usage.TotalTokens = 0
			usage.ResetAt = resetAt
			usage.LastUpdated = now
		}
	}
}

// Close stops the auto-reset goroutine
func (m *memoryQuotaManager) Close() error {
	if m.config.ResetPeriod != Never {
		close(m.stopReset)
	}
	return nil
}
