package config

import "time"

// Config holds all configuration for the system monitor
type Config struct {
	Interval   time.Duration // Refresh interval for metrics collection
	JSONMode   bool          // Enable JSON output mode
	LogFile    string        // Path to log file (empty if logging disabled)
	ConfigFile string        // Path to configuration file
	Thresholds Thresholds    // Alert thresholds
}

// Thresholds defines alert thresholds for different metrics
type Thresholds struct {
	CPU    float64 // CPU usage threshold (0-100)
	Memory float64 // Memory usage threshold (0-100)
	Disk   float64 // Disk usage threshold (0-100)
}

// NewDefaultConfig returns a Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		Interval: 1 * time.Second,
		JSONMode: false,
		LogFile:  "",
		Thresholds: Thresholds{
			CPU:    80.0,
			Memory: 85.0,
			Disk:   90.0,
		},
	}
}
