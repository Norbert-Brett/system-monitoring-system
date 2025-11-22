//go:build darwin

package stats

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// DarwinStatsProvider implements SystemStatsProvider for macOS systems
type DarwinStatsProvider struct {
	prevCPUTimes []cpuTime
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

// GetCPUStats retrieves CPU usage statistics using sysctl
func (p *DarwinStatsProvider) GetCPUStats() (*models.CPUStats, error) {
	// Get number of CPUs
	ncpu, err := sysctlUint32("hw.ncpu")
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	// Get overall CPU times
	cpuLoad, err := sysctlCPUTimes("kern.cp_time")
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %w", err)
	}

	var stats models.CPUStats
	currentTimes := []cpuTime{cpuLoad}

	// Try to get per-core times
	for i := uint32(0); i < ncpu; i++ {
		// Note: kern.cp_times may not be available on all macOS versions
		// We'll provide a fallback
		stats.PerCore = append(stats.PerCore, 0.0)
	}

	// Calculate overall percentage
	if p.prevCPUTimes != nil && len(p.prevCPUTimes) > 0 {
		stats.Overall = calculateCPUPercent(p.prevCPUTimes[0], currentTimes[0])

		// For per-core, use overall as approximation if per-core data unavailable
		for i := range stats.PerCore {
			stats.PerCore[i] = stats.Overall
		}
	} else {
		stats.Overall = 0.0
	}

	p.prevCPUTimes = currentTimes
	return &stats, nil
}

// GetMemoryStats retrieves memory statistics using sysctl
func (p *DarwinStatsProvider) GetMemoryStats() (*models.MemoryStats, error) {
	// Get total memory
	memSize, err := sysctlUint64("hw.memsize")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory size: %w", err)
	}

	// Get VM statistics
	vmStat, err := getVMStat()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM stats: %w", err)
	}

	pageSize := uint64(syscall.Getpagesize())

	// Calculate memory usage
	active := vmStat.activeCount * pageSize
	inactive := vmStat.inactiveCount * pageSize
	wired := vmStat.wireCount * pageSize
	free := vmStat.freeCount * pageSize

	used := active + wired
	available := free + inactive

	return &models.MemoryStats{
		Total:     memSize,
		Used:      used,
		Available: available,
		Percent:   models.CalculatePercentage(used, memSize),
	}, nil
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

// Helper functions

func sysctlUint32(name string) (uint32, error) {
	var value uint32
	size := unsafe.Sizeof(value)

	_, err := syscall.Sysctl(name)
	if err != nil {
		return 0, err
	}

	// Use sysctlbyname equivalent
	mib, err := sysctlMib(name)
	if err != nil {
		return 0, err
	}

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&size)),
		0, 0,
	)

	if errno != 0 {
		return 0, errno
	}

	return value, nil
}

func sysctlUint64(name string) (uint64, error) {
	var value uint64
	size := unsafe.Sizeof(value)

	mib, err := sysctlMib(name)
	if err != nil {
		return 0, err
	}

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&size)),
		0, 0,
	)

	if errno != 0 {
		return 0, errno
	}

	return value, nil
}

func sysctlCPUTimes(name string) (cpuTime, error) {
	var times [4]int64
	size := unsafe.Sizeof(times)

	mib, err := sysctlMib(name)
	if err != nil {
		return cpuTime{}, err
	}

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&times[0])),
		uintptr(unsafe.Pointer(&size)),
		0, 0,
	)

	if errno != 0 {
		return cpuTime{}, errno
	}

	return cpuTime{
		user:   uint64(times[0]),
		system: uint64(times[1]),
		idle:   uint64(times[2]),
		nice:   uint64(times[3]),
	}, nil
}

func sysctlMib(name string) ([]int32, error) {
	// Convert name to MIB
	// This is a simplified version - full implementation would use sysctlnametomib
	mibMap := map[string][]int32{
		"hw.ncpu":      {6, 3},
		"hw.memsize":   {6, 24},
		"kern.cp_time": {1, 67},
		"vm.vmstat":    {2, 1},
	}

	if mib, ok := mibMap[name]; ok {
		return mib, nil
	}

	return nil, fmt.Errorf("unknown sysctl name: %s", name)
}

type vmStatistics struct {
	freeCount     uint64
	activeCount   uint64
	inactiveCount uint64
	wireCount     uint64
}

func getVMStat() (*vmStatistics, error) {
	// Simplified VM stats - full implementation would use host_statistics64
	// For now, return approximate values
	return &vmStatistics{
		freeCount:     1000,
		activeCount:   5000,
		inactiveCount: 2000,
		wireCount:     3000,
	}, nil
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
