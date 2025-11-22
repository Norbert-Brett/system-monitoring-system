//go:build linux

package stats

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// LinuxStatsProvider implements SystemStatsProvider for Linux systems
type LinuxStatsProvider struct {
	prevCPUTimes []cpuTime
}

type cpuTime struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
}

// NewLinuxStatsProvider creates a new Linux stats provider
func NewLinuxStatsProvider() *LinuxStatsProvider {
	return &LinuxStatsProvider{}
}

// GetCPUStats retrieves CPU usage statistics from /proc/stat
func (p *LinuxStatsProvider) GetCPUStats() (*models.CPUStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/stat: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentTimes []cpuTime
	var stats models.CPUStats

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		// Parse CPU times
		times := cpuTime{
			user:    parseUint64(fields[1]),
			nice:    parseUint64(fields[2]),
			system:  parseUint64(fields[3]),
			idle:    parseUint64(fields[4]),
			iowait:  parseUint64(fields[5]),
			irq:     parseUint64(fields[6]),
			softirq: parseUint64(fields[7]),
		}

		currentTimes = append(currentTimes, times)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/stat: %w", err)
	}

	if len(currentTimes) == 0 {
		return nil, fmt.Errorf("no CPU data found in /proc/stat")
	}

	// Calculate percentages
	if p.prevCPUTimes != nil && len(p.prevCPUTimes) == len(currentTimes) {
		// Overall CPU (first entry)
		stats.Overall = calculateCPUPercent(p.prevCPUTimes[0], currentTimes[0])

		// Per-core CPU (remaining entries)
		for i := 1; i < len(currentTimes); i++ {
			percent := calculateCPUPercent(p.prevCPUTimes[i], currentTimes[i])
			stats.PerCore = append(stats.PerCore, percent)
		}
	} else {
		// First run - return zeros
		stats.Overall = 0.0
		for i := 1; i < len(currentTimes); i++ {
			stats.PerCore = append(stats.PerCore, 0.0)
		}
	}

	p.prevCPUTimes = currentTimes
	return &stats, nil
}

// GetMemoryStats retrieves memory statistics from /proc/meminfo
func (p *LinuxStatsProvider) GetMemoryStats() (*models.MemoryStats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	var memTotal, memFree, memAvailable, buffers, cached uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value := parseUint64(fields[1]) * 1024 // Convert KB to bytes

		switch key {
		case "MemTotal":
			memTotal = value
		case "MemFree":
			memFree = value
		case "MemAvailable":
			memAvailable = value
		case "Buffers":
			buffers = value
		case "Cached":
			cached = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/meminfo: %w", err)
	}

	// Use MemAvailable if present, otherwise calculate
	available := memAvailable
	if available == 0 {
		available = memFree + buffers + cached
	}

	used := memTotal - available

	return &models.MemoryStats{
		Total:     memTotal,
		Used:      used,
		Available: available,
		Percent:   models.CalculatePercentage(used, memTotal),
	}, nil
}

// GetDiskStats retrieves disk usage statistics using syscall.Statfs
func (p *LinuxStatsProvider) GetDiskStats() ([]models.DiskStats, error) {
	// Read mounted filesystems from /proc/mounts
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer file.Close()

	var stats []models.DiskStats
	scanner := bufio.NewScanner(file)
	seen := make(map[string]bool)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		mountpoint := fields[1]
		fstype := fields[2]

		// Skip special filesystems
		if strings.HasPrefix(fstype, "tmpfs") || strings.HasPrefix(fstype, "devtmpfs") ||
			strings.HasPrefix(fstype, "proc") || strings.HasPrefix(fstype, "sysfs") ||
			strings.HasPrefix(fstype, "cgroup") || strings.HasPrefix(fstype, "devpts") {
			continue
		}

		// Skip duplicates
		if seen[mountpoint] {
			continue
		}
		seen[mountpoint] = true

		// Get disk stats
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountpoint, &stat); err != nil {
			continue // Skip filesystems we can't stat
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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/mounts: %w", err)
	}

	return stats, nil
}

// GetNetworkStats retrieves network I/O statistics from /proc/net/dev
func (p *LinuxStatsProvider) GetNetworkStats() ([]models.NetworkStats, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/net/dev: %w", err)
	}
	defer file.Close()

	var stats []models.NetworkStats
	scanner := bufio.NewScanner(file)

	// Skip header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}

		// Skip loopback
		if iface == "lo" {
			continue
		}

		stats = append(stats, models.NetworkStats{
			Interface: iface,
			BytesRecv: parseUint64(fields[0]),
			BytesSent: parseUint64(fields[8]),
			SendRate:  0, // Rates calculated by collector
			RecvRate:  0,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/net/dev: %w", err)
	}

	return stats, nil
}

// Helper functions

func parseUint64(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}

func calculateCPUPercent(prev, curr cpuTime) float64 {
	prevTotal := prev.user + prev.nice + prev.system + prev.idle + prev.iowait + prev.irq + prev.softirq
	currTotal := curr.user + curr.nice + curr.system + curr.idle + curr.iowait + curr.irq + curr.softirq

	prevIdle := prev.idle + prev.iowait
	currIdle := curr.idle + curr.iowait

	totalDelta := currTotal - prevTotal
	idleDelta := currIdle - prevIdle

	if totalDelta == 0 {
		return 0.0
	}

	return (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100.0
}
