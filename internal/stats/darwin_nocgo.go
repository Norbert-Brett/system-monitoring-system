//go:build darwin && !cgo

package stats

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sysmon/system-monitor-cli/internal/models"
	"golang.org/x/sys/unix"
)

func getDarwinMemoryStats() (*models.MemoryStats, error) {
	memSize, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory size: %w", err)
	}

	pageSize, err := unix.SysctlUint32("hw.pagesize")
	if err != nil || pageSize == 0 {
		pageSize = uint32(unix.Getpagesize())
	}
	ps := uint64(pageSize)

	out, err := exec.Command("/usr/bin/vm_stat").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run vm_stat: %w", err)
	}

	var active, wired, compressed uint64
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "."))
		val, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "Pages active":
			active = val
		case "Pages wired down":
			wired = val
		case "Pages occupied by compressor":
			compressed = val
		}
	}

	used := (active + wired + compressed) * ps
	if used > memSize {
		used = memSize
	}
	available := memSize - used

	return &models.MemoryStats{
		Total:     memSize,
		Used:      used,
		Available: available,
		Percent:   models.CalculatePercentage(used, memSize),
	}, nil
}

func getDarwinCPUStats(prevOverall *cpuTime, prevPerCore []cpuTime) (*models.CPUStats, *cpuTime, []cpuTime, error) {
	ncpu, err := unix.SysctlUint32("hw.ncpu")
	if err != nil {
		ncpu = 1
	}

	stats := &models.CPUStats{
		Overall: 0.0,
		PerCore: make([]float64, ncpu),
	}
	return stats, prevOverall, prevPerCore, nil
}
