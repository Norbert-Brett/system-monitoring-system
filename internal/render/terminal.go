package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sysmon/system-monitor-cli/internal/config"
	"github.com/sysmon/system-monitor-cli/internal/models"
	"golang.org/x/term"
)

// ANSI control codes
const (
	ansiClearScreen = "\033[2J"
	ansiHome        = "\033[H"
	ansiClearLine   = "\033[2K"
)

// TerminalRenderer renders metrics to a terminal with ANSI formatting
type TerminalRenderer struct {
	writer     io.Writer
	thresholds *config.Thresholds
	useANSI    bool
}

// NewTerminalRenderer creates a new terminal renderer
func NewTerminalRenderer(writer io.Writer, thresholds *config.Thresholds) *TerminalRenderer {
	// Detect if output supports ANSI codes
	useANSI := isTerminal(writer)

	return &TerminalRenderer{
		writer:     writer,
		thresholds: thresholds,
		useANSI:    useANSI,
	}
}

// Render formats and displays metrics in a terminal-friendly layout
func (r *TerminalRenderer) Render(metrics *models.Metrics) error {
	var output strings.Builder

	// Clear screen and move cursor to home if ANSI supported
	if r.useANSI {
		output.WriteString(ansiClearScreen)
		output.WriteString(ansiHome)
	} else {
		output.WriteString("\n" + strings.Repeat("=", 80) + "\n")
	}

	// Header
	output.WriteString(r.formatHeader(metrics))
	output.WriteString("\n")

	// CPU Section
	output.WriteString(r.formatCPU(metrics.CPU))
	output.WriteString("\n")

	// Memory Section
	output.WriteString(r.formatMemory(metrics.Memory))
	output.WriteString("\n")

	// Disk Section
	if len(metrics.Disk) > 0 {
		output.WriteString(r.formatDisk(metrics.Disk))
		output.WriteString("\n")
	}

	// Network Section
	if len(metrics.Network) > 0 {
		output.WriteString(r.formatNetwork(metrics.Network))
		output.WriteString("\n")
	}

	_, err := r.writer.Write([]byte(output.String()))
	return err
}

// Clear clears the terminal display
func (r *TerminalRenderer) Clear() error {
	if r.useANSI {
		_, err := r.writer.Write([]byte(ansiClearScreen + ansiHome))
		return err
	}
	return nil
}

// Close performs cleanup
func (r *TerminalRenderer) Close() error {
	// Restore terminal
	if r.useANSI {
		_, err := r.writer.Write([]byte("\n"))
		return err
	}
	return nil
}

// formatHeader creates the header section
func (r *TerminalRenderer) formatHeader(metrics *models.Metrics) string {
	title := "System Monitor"
	timestamp := metrics.Timestamp.Format("2006-01-02 15:04:05")

	if r.useANSI {
		titleColor := color.New(color.FgCyan, color.Bold)
		return fmt.Sprintf("%s - %s\n", titleColor.Sprint(title), timestamp)
	}
	return fmt.Sprintf("%s - %s\n", title, timestamp)
}

// formatCPU formats CPU statistics
func (r *TerminalRenderer) formatCPU(cpu models.CPUStats) string {
	var output strings.Builder

	// Section header
	if r.useANSI {
		header := color.New(color.FgYellow, color.Bold).Sprint("CPU Usage:")
		output.WriteString(header + "\n")
	} else {
		output.WriteString("CPU Usage:\n")
	}

	// Overall CPU
	warning := r.shouldWarn(cpu.Overall, r.thresholds.CPU)
	overallStr := fmt.Sprintf("  Overall: %6.2f%%", cpu.Overall)
	if warning {
		overallStr += " " + r.formatWarning()
	}
	output.WriteString(r.colorizeValue(overallStr, cpu.Overall, r.thresholds.CPU) + "\n")

	// Per-core CPU
	if len(cpu.PerCore) > 0 {
		output.WriteString("  Per Core:\n")
		for i, percent := range cpu.PerCore {
			coreStr := fmt.Sprintf("    Core %2d: %6.2f%%", i, percent)
			warning := r.shouldWarn(percent, r.thresholds.CPU)
			if warning {
				coreStr += " " + r.formatWarning()
			}
			output.WriteString(r.colorizeValue(coreStr, percent, r.thresholds.CPU) + "\n")
		}
	}

	return output.String()
}

// formatMemory formats memory statistics
func (r *TerminalRenderer) formatMemory(mem models.MemoryStats) string {
	var output strings.Builder

	// Section header
	if r.useANSI {
		header := color.New(color.FgYellow, color.Bold).Sprint("Memory Usage:")
		output.WriteString(header + "\n")
	} else {
		output.WriteString("Memory Usage:\n")
	}

	// Memory stats
	totalGB := float64(mem.Total) / (1024 * 1024 * 1024)
	usedGB := float64(mem.Used) / (1024 * 1024 * 1024)
	availGB := float64(mem.Available) / (1024 * 1024 * 1024)

	warning := r.shouldWarn(mem.Percent, r.thresholds.Memory)
	percentStr := fmt.Sprintf("  Usage:     %6.2f%%", mem.Percent)
	if warning {
		percentStr += " " + r.formatWarning()
	}
	output.WriteString(r.colorizeValue(percentStr, mem.Percent, r.thresholds.Memory) + "\n")

	output.WriteString(fmt.Sprintf("  Total:     %8.2f GB\n", totalGB))
	output.WriteString(fmt.Sprintf("  Used:      %8.2f GB\n", usedGB))
	output.WriteString(fmt.Sprintf("  Available: %8.2f GB\n", availGB))

	return output.String()
}

// formatDisk formats disk statistics
func (r *TerminalRenderer) formatDisk(disks []models.DiskStats) string {
	var output strings.Builder

	// Section header
	if r.useANSI {
		header := color.New(color.FgYellow, color.Bold).Sprint("Disk Usage:")
		output.WriteString(header + "\n")
	} else {
		output.WriteString("Disk Usage:\n")
	}

	for _, disk := range disks {
		totalGB := float64(disk.Total) / (1024 * 1024 * 1024)
		usedGB := float64(disk.Used) / (1024 * 1024 * 1024)
		availGB := float64(disk.Available) / (1024 * 1024 * 1024)

		output.WriteString(fmt.Sprintf("  %s\n", disk.Mountpoint))

		warning := r.shouldWarn(disk.Percent, r.thresholds.Disk)
		percentStr := fmt.Sprintf("    Usage:     %6.2f%%", disk.Percent)
		if warning {
			percentStr += " " + r.formatWarning()
		}
		output.WriteString(r.colorizeValue(percentStr, disk.Percent, r.thresholds.Disk) + "\n")

		output.WriteString(fmt.Sprintf("    Total:     %8.2f GB\n", totalGB))
		output.WriteString(fmt.Sprintf("    Used:      %8.2f GB\n", usedGB))
		output.WriteString(fmt.Sprintf("    Available: %8.2f GB\n", availGB))
	}

	return output.String()
}

// formatNetwork formats network statistics
func (r *TerminalRenderer) formatNetwork(networks []models.NetworkStats) string {
	var output strings.Builder

	// Section header
	if r.useANSI {
		header := color.New(color.FgYellow, color.Bold).Sprint("Network I/O:")
		output.WriteString(header + "\n")
	} else {
		output.WriteString("Network I/O:\n")
	}

	for _, net := range networks {
		output.WriteString(fmt.Sprintf("  %s\n", net.Interface))
		output.WriteString(fmt.Sprintf("    Sent:     %s (%s/s)\n",
			formatBytes(net.BytesSent), formatBytes(uint64(net.SendRate))))
		output.WriteString(fmt.Sprintf("    Received: %s (%s/s)\n",
			formatBytes(net.BytesRecv), formatBytes(uint64(net.RecvRate))))
	}

	return output.String()
}

// colorizeValue applies color based on threshold
func (r *TerminalRenderer) colorizeValue(text string, value, threshold float64) string {
	if !r.useANSI {
		return text
	}

	if value > threshold {
		return color.RedString(text)
	} else if value > threshold*0.8 {
		return color.YellowString(text)
	}
	return color.GreenString(text)
}

// shouldWarn checks if a value exceeds the threshold
func (r *TerminalRenderer) shouldWarn(value, threshold float64) bool {
	return value > threshold
}

// formatWarning returns a warning indicator
func (r *TerminalRenderer) formatWarning() string {
	if r.useANSI {
		return color.RedString("⚠ WARNING")
	}
	return "[WARNING]"
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
