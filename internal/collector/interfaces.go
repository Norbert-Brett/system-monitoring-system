package collector

import (
	"context"
	"time"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// MetricsCollector defines the interface for collecting system metrics
type MetricsCollector interface {
	// Collect gathers a single snapshot of system metrics
	Collect(ctx context.Context) (*models.Metrics, error)

	// Start begins periodic metric collection, sending results to the output channel
	Start(ctx context.Context, interval time.Duration, out chan<- *models.Metrics) error
}

// SystemStatsProvider defines the interface for OS-specific system statistics
type SystemStatsProvider interface {
	// GetCPUStats retrieves CPU usage statistics
	GetCPUStats() (*models.CPUStats, error)

	// GetMemoryStats retrieves memory usage statistics
	GetMemoryStats() (*models.MemoryStats, error)

	// GetDiskStats retrieves disk usage statistics for all mounted filesystems
	GetDiskStats() ([]models.DiskStats, error)

	// GetNetworkStats retrieves network I/O statistics for all interfaces
	GetNetworkStats() ([]models.NetworkStats, error)
}
