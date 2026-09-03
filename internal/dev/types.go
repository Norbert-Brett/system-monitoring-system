package dev

import (
	"time"
)

// ListeningPort represents a network port in LISTEN state
type ListeningPort struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	User        string `json:"user"`
	CWD         string `json:"cwd,omitempty"`
	Project     string `json:"project,omitempty"`
	Category    string `json:"category,omitempty"`
	Stack       string `json:"stack,omitempty"`
}

// DevProcess represents a developer-relevant process
type DevProcess struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Stack      string  `json:"stack"`
	CWD        string  `json:"cwd,omitempty"`
	Project    string  `json:"project,omitempty"`
	RSSBytes   uint64  `json:"rss_bytes"`
	CPUPercent float64 `json:"cpu_percent"`
}

// CommandProfile holds benchmark and resource metrics for a command
type CommandProfile struct {
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	PeakRSS   uint64        `json:"peak_rss_bytes"`
	UserCPU   time.Duration `json:"user_cpu_time"`
	SysCPU    time.Duration `json:"sys_cpu_time"`
	Timestamp time.Time     `json:"timestamp"`
}

// SnapshotReport contains an aggregated diagnostic summary of the dev environment
type SnapshotReport struct {
	Timestamp      time.Time         `json:"timestamp"`
	OS             string            `json:"os"`
	Kernel         string            `json:"kernel"`
	Arch           string            `json:"arch"`
	CPUCores       int               `json:"cpu_cores"`
	MemoryTotal    uint64            `json:"memory_total_bytes"`
	MemoryUsed     uint64            `json:"memory_used_bytes"`
	MemoryAvail    uint64            `json:"memory_avail_bytes"`
	MemoryPercent  float64           `json:"memory_percent"`
	Runtimes       map[string]string `json:"runtimes"`
	ListeningPorts []ListeningPort   `json:"listening_ports"`
}
