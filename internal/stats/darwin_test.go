//go:build darwin

package stats

import (
	"testing"
	"time"
)

func TestDarwinMemoryStats(t *testing.T) {
	provider := NewDarwinStatsProvider()
	mem, err := provider.GetMemoryStats()
	if err != nil {
		t.Fatalf("GetMemoryStats() failed: %v", err)
	}

	if mem.Total == 0 {
		t.Errorf("expected Total > 0, got %d", mem.Total)
	}
	if mem.Used == 0 {
		t.Errorf("expected Used > 0, got %d", mem.Used)
	}
	if mem.Available == 0 {
		t.Errorf("expected Available > 0, got %d", mem.Available)
	}
	if mem.Used > mem.Total {
		t.Errorf("expected Used (%d) <= Total (%d)", mem.Used, mem.Total)
	}
	if mem.Available > mem.Total {
		t.Errorf("expected Available (%d) <= Total (%d)", mem.Available, mem.Total)
	}
	if mem.Percent <= 0 || mem.Percent > 100 {
		t.Errorf("expected Percent between 0 and 100, got %f", mem.Percent)
	}

	// Verify Used + Available accounts for Total (within small rounding/rounding of total - used)
	if mem.Used+mem.Available != mem.Total {
		t.Errorf("expected Used + Available == Total, got %d + %d != %d", mem.Used, mem.Available, mem.Total)
	}

	t.Logf("Darwin Memory: Total=%.2f GB, Used=%.2f GB, Available=%.2f GB, Percent=%.2f%%",
		float64(mem.Total)/(1024*1024*1024),
		float64(mem.Used)/(1024*1024*1024),
		float64(mem.Available)/(1024*1024*1024),
		mem.Percent,
	)
}

func TestDarwinCPUStats(t *testing.T) {
	provider := NewDarwinStatsProvider()

	// First call initializes previous counters
	stats1, err := provider.GetCPUStats()
	if err != nil {
		t.Fatalf("First GetCPUStats() failed: %v", err)
	}
	if stats1 == nil {
		t.Fatal("First GetCPUStats() returned nil")
	}

	// Sleep briefly so there are deltas
	time.Sleep(100 * time.Millisecond)

	// Second call calculates percentage deltas
	stats2, err := provider.GetCPUStats()
	if err != nil {
		t.Fatalf("Second GetCPUStats() failed: %v", err)
	}

	if stats2.Overall < 0 || stats2.Overall > 100 {
		t.Errorf("expected Overall CPU between 0 and 100, got %f", stats2.Overall)
	}
	if len(stats2.PerCore) == 0 {
		t.Error("expected at least 1 core in PerCore")
	}

	for i, core := range stats2.PerCore {
		if core < 0 || core > 100 {
			t.Errorf("core %d: expected between 0 and 100, got %f", i, core)
		}
	}

	t.Logf("Darwin CPU: Overall=%.2f%%, Cores=%d", stats2.Overall, len(stats2.PerCore))
}
