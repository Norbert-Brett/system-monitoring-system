package collector

import (
	"context"
	"log"
	"time"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// Collector implements MetricsCollector using a SystemStatsProvider
type Collector struct {
	provider SystemStatsProvider
	prevNet  []models.NetworkStats
	prevTime time.Time
}

// NewCollector creates a new metrics collector with the given provider
func NewCollector(provider SystemStatsProvider) *Collector {
	return &Collector{
		provider: provider,
		prevTime: time.Now(),
	}
}

// Collect gathers a single snapshot of system metrics
// It implements partial failure handling - errors in individual subsystems
// don't prevent collection of other metrics
func (c *Collector) Collect(ctx context.Context) (*models.Metrics, error) {
	metrics := &models.Metrics{
		Timestamp: time.Now(),
	}

	// Collect CPU stats
	if cpu, err := c.provider.GetCPUStats(); err != nil {
		log.Printf("Warning: CPU collection failed: %v", err)
	} else {
		metrics.CPU = *cpu
	}

	// Collect memory stats
	if mem, err := c.provider.GetMemoryStats(); err != nil {
		log.Printf("Warning: Memory collection failed: %v", err)
	} else {
		metrics.Memory = *mem
	}

	// Collect disk stats
	if disk, err := c.provider.GetDiskStats(); err != nil {
		log.Printf("Warning: Disk collection failed: %v", err)
	} else {
		metrics.Disk = disk
	}

	// Collect network stats and calculate rates
	if net, err := c.provider.GetNetworkStats(); err != nil {
		log.Printf("Warning: Network collection failed: %v", err)
	} else {
		// Calculate rates if we have previous data
		if c.prevNet != nil {
			net = c.calculateNetworkRates(net)
		}
		metrics.Network = net
		c.prevNet = net
		c.prevTime = metrics.Timestamp
	}

	return metrics, nil
}

// Start begins periodic metric collection, sending results to the output channel
// It runs in a goroutine and respects context cancellation
func (c *Collector) Start(ctx context.Context, interval time.Duration, out chan<- *models.Metrics) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(out)

	// Collect initial metrics immediately
	if metrics, err := c.Collect(ctx); err == nil {
		select {
		case out <- metrics:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			metrics, err := c.Collect(ctx)
			if err != nil {
				log.Printf("Error collecting metrics: %v", err)
				continue
			}

			select {
			case out <- metrics:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// calculateNetworkRates computes send and receive rates based on previous measurements
func (c *Collector) calculateNetworkRates(current []models.NetworkStats) []models.NetworkStats {
	if c.prevNet == nil || len(c.prevNet) == 0 {
		return current
	}

	timeDelta := time.Since(c.prevTime).Seconds()
	if timeDelta == 0 {
		return current
	}

	// Create a map of previous stats by interface name for quick lookup
	prevMap := make(map[string]models.NetworkStats)
	for _, prev := range c.prevNet {
		prevMap[prev.Interface] = prev
	}

	// Calculate rates for each interface
	result := make([]models.NetworkStats, len(current))
	for i, curr := range current {
		result[i] = curr

		if prev, exists := prevMap[curr.Interface]; exists {
			// Calculate bytes per second
			bytesSentDelta := float64(curr.BytesSent - prev.BytesSent)
			bytesRecvDelta := float64(curr.BytesRecv - prev.BytesRecv)

			result[i].SendRate = bytesSentDelta / timeDelta
			result[i].RecvRate = bytesRecvDelta / timeDelta

			// Handle counter wraps (unlikely but possible)
			if result[i].SendRate < 0 {
				result[i].SendRate = 0
			}
			if result[i].RecvRate < 0 {
				result[i].RecvRate = 0
			}
		}
	}

	return result
}
