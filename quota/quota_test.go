package quota

import (
	"context"
	"testing"
	"time"
)

func TestQuotaManager_RecordAndCheck(t *testing.T) {
	config := &Config{
		DefaultQuota: 1000,
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage
	err := mgr.RecordUsage(ctx, "tenant1", 100, 50, 150)
	if err != nil {
		t.Fatalf("Failed to record usage: %v", err)
	}
	
	// Check quota
	hasQuota, usage, err := mgr.CheckQuota(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	
	if !hasQuota {
		t.Error("Expected tenant to have quota remaining")
	}
	
	if usage.TotalTokens != 150 {
		t.Errorf("Expected 150 total tokens, got %d", usage.TotalTokens)
	}
}

func TestQuotaManager_ExceedQuota(t *testing.T) {
	config := &Config{
		DefaultQuota: 100,
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage that exceeds quota
	mgr.RecordUsage(ctx, "tenant1", 150, 0, 150)
	
	// Check quota - should be exceeded
	hasQuota, usage, err := mgr.CheckQuota(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	
	if hasQuota {
		t.Errorf("Expected tenant to have exceeded quota (150 > 100), but hasQuota=%v", hasQuota)
	}
	
	if usage.TotalTokens != 150 {
		t.Errorf("Expected 150 total tokens, got %d", usage.TotalTokens)
	}
}

func TestQuotaManager_UnlimitedQuota(t *testing.T) {
	config := &Config{
		DefaultQuota: 0, // Unlimited
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record large usage
	mgr.RecordUsage(ctx, "tenant1", 10000, 10000, 20000)
	
	// Check quota - should still have quota (unlimited)
	hasQuota, _, err := mgr.CheckQuota(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	
	if !hasQuota {
		t.Error("Expected tenant to have quota (unlimited)")
	}
}

func TestQuotaManager_SetQuota(t *testing.T) {
	config := &Config{
		DefaultQuota: 100,
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Set custom quota
	err := mgr.SetQuota(ctx, "tenant1", 500)
	if err != nil {
		t.Fatalf("Failed to set quota: %v", err)
	}
	
	// Record usage
	mgr.RecordUsage(ctx, "tenant1", 300, 0, 300)
	
	// Should still have quota
	hasQuota, usage, err := mgr.CheckQuota(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	
	if !hasQuota {
		t.Error("Expected tenant to have quota")
	}
	
	if usage.QuotaLimit != 500 {
		t.Errorf("Expected quota limit of 500, got %d", usage.QuotaLimit)
	}
}

func TestQuotaManager_ResetUsage(t *testing.T) {
	config := &Config{
		DefaultQuota: 1000,
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage
	mgr.RecordUsage(ctx, "tenant1", 500, 500, 1000)
	
	// Reset usage
	err := mgr.ResetUsage(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to reset usage: %v", err)
	}
	
	// Check usage
	usage, err := mgr.GetUsage(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to get usage: %v", err)
	}
	
	if usage.TotalTokens != 0 {
		t.Errorf("Expected 0 total tokens after reset, got %d", usage.TotalTokens)
	}
}

func TestQuotaManager_MultipleTenantsIsolation(t *testing.T) {
	config := &Config{
		DefaultQuota: 1000,
		ResetPeriod:  Never,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage for different tenants
	mgr.RecordUsage(ctx, "tenant1", 100, 0, 100)
	mgr.RecordUsage(ctx, "tenant2", 200, 0, 200)
	
	// Check tenant1
	usage1, _ := mgr.GetUsage(ctx, "tenant1")
	if usage1.TotalTokens != 100 {
		t.Errorf("Expected tenant1 to have 100 tokens, got %d", usage1.TotalTokens)
	}
	
	// Check tenant2
	usage2, _ := mgr.GetUsage(ctx, "tenant2")
	if usage2.TotalTokens != 200 {
		t.Errorf("Expected tenant2 to have 200 tokens, got %d", usage2.TotalTokens)
	}
}

func TestQuotaManager_Disabled(t *testing.T) {
	config := &Config{
		DefaultQuota: 100,
		ResetPeriod:  Never,
		Enabled:      false, // Disabled
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage (should be no-op)
	mgr.RecordUsage(ctx, "tenant1", 1000, 1000, 2000)
	
	// Check quota (should always return true when disabled)
	hasQuota, _, err := mgr.CheckQuota(ctx, "tenant1")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	
	if !hasQuota {
		t.Error("Expected quota check to return true when disabled")
	}
}

func TestQuotaManager_ResetPeriod(t *testing.T) {
	config := &Config{
		DefaultQuota: 1000,
		ResetPeriod:  Hourly,
		Enabled:      true,
	}
	
	mgr := NewMemoryManager(config)
	ctx := context.Background()
	
	// Record usage
	mgr.RecordUsage(ctx, "tenant1", 500, 0, 500)
	
	// Get usage
	usage, _ := mgr.GetUsage(ctx, "tenant1")
	
	// Check that reset time is set
	if usage.ResetAt.IsZero() {
		t.Error("Expected ResetAt to be set")
	}
	
	// Reset time should be in the future
	if !usage.ResetAt.After(time.Now()) {
		t.Error("Expected ResetAt to be in the future")
	}
}
