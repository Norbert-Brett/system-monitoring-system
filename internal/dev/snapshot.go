package dev

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sysmon/system-monitor-cli/internal/stats"
)

// GenerateSnapshot creates a diagnostic snapshot report of the system and developer runtimes
func GenerateSnapshot() (*SnapshotReport, error) {
	provider, err := stats.NewProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stats provider: %w", err)
	}

	mem, err := provider.GetMemoryStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	ports, _ := GetListeningPorts()

	// Detect dev runtimes
	runtimes := make(map[string]string)
	tools := map[string][]string{
		"Go":      {"go", "version"},
		"Node.js": {"node", "--version"},
		"npm":     {"npm", "--version"},
		"Python":  {"python3", "--version"},
		"Docker":  {"docker", "--version"},
		"Git":     {"git", "--version"},
		"Rust":    {"rustc", "--version"},
	}

	for tool, args := range tools {
		out, err := exec.Command(args[0], args[1:]...).Output()
		if err == nil {
			val := strings.TrimSpace(string(out))
			if strings.Contains(val, "\n") {
				val = strings.Split(val, "\n")[0]
			}
			runtimes[tool] = val
		}
	}

	kernel := "unknown"
	unameOut, err := exec.Command("uname", "-r").Output()
	if err == nil {
		kernel = strings.TrimSpace(string(unameOut))
	}

	return &SnapshotReport{
		Timestamp:      time.Now(),
		OS:             runtime.GOOS,
		Kernel:         kernel,
		Arch:           runtime.GOARCH,
		CPUCores:       runtime.NumCPU(),
		MemoryTotal:    mem.Total,
		MemoryUsed:     mem.Used,
		MemoryAvail:    mem.Available,
		MemoryPercent:  mem.Percent,
		Runtimes:       runtimes,
		ListeningPorts: ports,
	}, nil
}

// FormatMarkdown formats the snapshot report into clean Markdown
func (s *SnapshotReport) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# 🔍 Developer Environment & System Snapshot\n\n")
	sb.WriteString(fmt.Sprintf("**Generated At**: %s\n\n", s.Timestamp.Format("2006-01-02 15:04:05 MST")))

	sb.WriteString("## 💻 Host System\n\n")
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **OS / Arch** | `%s/%s` |\n", s.OS, s.Arch))
	sb.WriteString(fmt.Sprintf("| **Kernel** | `%s` |\n", s.Kernel))
	sb.WriteString(fmt.Sprintf("| **CPU Cores** | `%d` |\n", s.CPUCores))
	sb.WriteString(fmt.Sprintf("| **RAM Total** | `%.2f GB` |\n", float64(s.MemoryTotal)/(1024*1024*1024)))
	sb.WriteString(fmt.Sprintf("| **RAM Used** | `%.2f GB (%.1f%%)` |\n", float64(s.MemoryUsed)/(1024*1024*1024), s.MemoryPercent))
	sb.WriteString(fmt.Sprintf("| **RAM Available** | `%.2f GB` |\n\n", float64(s.MemoryAvail)/(1024*1024*1024)))

	sb.WriteString("## 🛠️ Developer Runtimes\n\n")
	if len(s.Runtimes) == 0 {
		sb.WriteString("*(No standard developer runtimes found in PATH)*\n\n")
	} else {
		sb.WriteString("| Tool | Version |\n")
		sb.WriteString("|------|---------|\n")
		for tool, ver := range s.Runtimes {
			sb.WriteString(fmt.Sprintf("| **%s** | `%s` |\n", tool, ver))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 🔌 Active Listening Ports\n\n")
	if len(s.ListeningPorts) == 0 {
		sb.WriteString("*(No active listening TCP ports detected)*\n")
	} else {
		sb.WriteString("| Port | Process | PID | Stack / Category | Project / CWD |\n")
		sb.WriteString("|------|---------|-----|------------------|---------------|\n")
		for _, p := range s.ListeningPorts {
			loc := p.Project
			if loc == "" {
				loc = p.CWD
			}
			if loc == "" {
				loc = "-"
			}
			stackCat := p.Stack
			if p.Category != "" && p.Category != "System" {
				stackCat += " (" + p.Category + ")"
			}
			if stackCat == "" {
				stackCat = "System"
			}
			sb.WriteString(fmt.Sprintf("| `:%d` | **%s** | `%d` | %s | `%s` |\n",
				p.Port, p.ProcessName, p.PID, stackCat, loc))
		}
	}

	return sb.String()
}
