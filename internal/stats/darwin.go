//go:build darwin

package stats

import (
	"fmt"
	"syscall"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// DarwinStatsProvider implements SystemStatsProvider for macOS systems
type DarwinStatsProvider struct {
	prevOverall *cpuTime
	prevPerCore []cpuTime
}

type cpuTime struct {
	user   uint64
	system uint64
	idle   uint64
	nice   uint64
}

// NewDarwinStatsProvider creates a new Darwin stats provider
func NewDarwinStatsProvider() *DarwinStatsProvider {
	return &DarwinStatsProvider{}
}

// GetCPUStats retrieves CPU usage statistics
func (p *DarwinStatsProvider) GetCPUStats() (*models.CPUStats, error) {
	stats, currOverall, currPerCore, err := getDarwinCPUStats(p.prevOverall, p.prevPerCore)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU stats: %w", err)
	}

	p.prevOverall = currOverall
	p.prevPerCore = currPerCore
	return stats, nil
}

// GetMemoryStats retrieves memory statistics
func (p *DarwinStatsProvider) GetMemoryStats() (*models.MemoryStats, error) {
	return getDarwinMemoryStats()
}

// GetDiskStats retrieves disk usage statistics using syscall.Statfs
func (p *DarwinStatsProvider) GetDiskStats() ([]models.DiskStats, error) {
	// Get mounted filesystems
	var stats []models.DiskStats

	// Common mount points on macOS
	mountpoints := []string{"/", "/System/Volumes/Data"}

	for _, mountpoint := range mountpoints {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountpoint, &stat); err != nil {
			continue // Skip if we can't stat
		}

		total := stat.Blocks * uint64(stat.Bsize)
		available := stat.Bavail * uint64(stat.Bsize)
		used := total - (stat.Bfree * uint64(stat.Bsize))

		stats = append(stats, models.DiskStats{
			Mountpoint: mountpoint,
			Total:      total,
			Used:       used,
			Available:  available,
			Percent:    models.CalculatePercentage(used, total),
		})
	}

	if len(stats) == 0 {
		return nil, fmt.Errorf("no disk stats available")
	}

	return stats, nil
}

// GetNetworkStats retrieves network I/O statistics
func (p *DarwinStatsProvider) GetNetworkStats() ([]models.NetworkStats, error) {
	// Note: Getting network stats on macOS requires more complex syscalls
	// For now, return empty stats - this would need IOKit framework integration
	// or parsing netstat output for a complete implementation
	return []models.NetworkStats{}, nil
}

func calculateCPUPercent(prev, curr cpuTime) float64 {
	prevTotal := prev.user + prev.system + prev.idle + prev.nice
	currTotal := curr.user + curr.system + curr.idle + curr.nice

	totalDelta := currTotal - prevTotal
	idleDelta := curr.idle - prev.idle

	if totalDelta == 0 {
		return 0.0
	}

	return (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100.0
}
