package models

import "time"

// Metrics represents a complete snapshot of system metrics at a point in time
type Metrics struct {
	Timestamp time.Time
	CPU       CPUStats
	Memory    MemoryStats
	Disk      []DiskStats
	Network   []NetworkStats
}

// CPUStats represents CPU usage statistics
type CPUStats struct {
	Overall float64   // Overall CPU usage percentage (0-100)
	PerCore []float64 // Per-core usage percentages (0-100)
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Total     uint64  // Total memory in bytes
	Used      uint64  // Used memory in bytes
	Available uint64  // Available memory in bytes
	Percent   float64 // Usage percentage (0-100)
}

// DiskStats represents disk usage statistics for a filesystem
type DiskStats struct {
	Mountpoint string
	Total      uint64  // Total space in bytes
	Used       uint64  // Used space in bytes
	Available  uint64  // Available space in bytes
	Percent    float64 // Usage percentage (0-100)
}

// NetworkStats represents network I/O statistics for an interface
type NetworkStats struct {
	Interface string
	BytesSent uint64  // Total bytes sent
	BytesRecv uint64  // Total bytes received
	SendRate  float64 // Bytes per second
	RecvRate  float64 // Bytes per second
}

// CalculatePercentage calculates percentage from used and total values
func CalculatePercentage(used, total uint64) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(used) / float64(total)) * 100.0
}
