package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/sysmon/system-monitor-cli/internal/collector"
	"github.com/sysmon/system-monitor-cli/internal/config"
	"github.com/sysmon/system-monitor-cli/internal/logger"
	"github.com/sysmon/system-monitor-cli/internal/monitor"
	"github.com/sysmon/system-monitor-cli/internal/render"
	"github.com/sysmon/system-monitor-cli/internal/stats"
)

var (
	// Flag variables
	cfgFile       string
	interval      time.Duration
	jsonMode      bool
	logFile       string
	cpuThreshold  float64
	memThreshold  float64
	diskThreshold float64

	// Version information
	Version = "1.0.0"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "sysmon",
	Short: "System Monitor - Real-time system metrics monitoring",
	Long: `System Monitor is a command-line tool that displays real-time system information
including CPU usage, memory usage, disk usage, and network I/O statistics.

The tool provides both interactive terminal display with colorized output and
JSON output mode for integration with other tools.`,
	RunE: runMonitor,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (YAML or JSON)")
	rootCmd.PersistentFlags().DurationVar(&interval, "interval", 1*time.Second, "refresh interval (e.g., 1s, 500ms, 2m)")
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "output metrics as JSON")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "path to log file for metrics export")
	rootCmd.PersistentFlags().Float64Var(&cpuThreshold, "cpu-threshold", 80.0, "CPU usage alert threshold (0-100)")
	rootCmd.PersistentFlags().Float64Var(&memThreshold, "mem-threshold", 85.0, "memory usage alert threshold (0-100)")
	rootCmd.PersistentFlags().Float64Var(&diskThreshold, "disk-threshold", 90.0, "disk usage alert threshold (0-100)")
}

// runMonitor is the main execution function for the monitor command
func runMonitor(cmd *cobra.Command, args []string) error {
	// Load configuration from file if specified
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check which flags were explicitly set
	intervalSet := cmd.Flags().Changed("interval")
	jsonSet := cmd.Flags().Changed("json")
	logFileSet := cmd.Flags().Changed("log-file")
	cpuThresholdSet := cmd.Flags().Changed("cpu-threshold")
	memThresholdSet := cmd.Flags().Changed("mem-threshold")
	diskThresholdSet := cmd.Flags().Changed("disk-threshold")

	// Merge with command-line flags (flags take precedence)
	if intervalSet {
		cfg.Interval = interval
	}
	if jsonSet {
		cfg.JSONMode = jsonMode
	}
	if logFileSet {
		cfg.LogFile = logFile
	}
	if cpuThresholdSet {
		cfg.Thresholds.CPU = cpuThreshold
	}
	if memThresholdSet {
		cfg.Thresholds.Memory = memThreshold
	}
	if diskThresholdSet {
		cfg.Thresholds.Disk = diskThreshold
	}

	// Validate final configuration
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create system stats provider
	provider, err := stats.NewProvider()
	if err != nil {
		return fmt.Errorf("failed to create stats provider: %w", err)
	}

	// Create metrics collector
	metricsCollector := collector.NewCollector(provider)

	// Create renderer based on mode
	var renderer render.Renderer
	if cfg.JSONMode {
		renderer = render.NewJSONRenderer(os.Stdout)
	} else {
		renderer = render.NewTerminalRenderer(os.Stdout, &cfg.Thresholds)
	}

	// Create logger if log file specified
	var metricsLogger logger.Logger
	if cfg.LogFile != "" {
		metricsLogger, err = logger.NewFileLogger(cfg.LogFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create logger: %v\n", err)
			metricsLogger = nil
		}
	}

	// Create monitor
	mon := monitor.NewSystemMonitor(cfg, metricsCollector, renderer, metricsLogger)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start monitor in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- mon.Start(ctx)
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Fprintln(os.Stderr, "\nShutting down gracefully...")
		cancel()
		// Wait for monitor to stop
		<-errChan
		return mon.Stop()
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			return err
		}
		return mon.Stop()
	}
}
