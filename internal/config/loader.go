package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// LoadFromFile loads configuration from a file (YAML or JSON)
func LoadFromFile(path string) (*Config, error) {
	if path == "" {
		return NewDefaultConfig(), nil
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := NewDefaultConfig()

	// Load interval
	if v.IsSet("interval") {
		intervalStr := v.GetString("interval")
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid interval format: %w", err)
		}
		config.Interval = interval
	}

	// Load JSON mode
	if v.IsSet("json") {
		config.JSONMode = v.GetBool("json")
	}

	// Load log file
	if v.IsSet("logFile") {
		config.LogFile = v.GetString("logFile")
	}

	// Load thresholds
	if v.IsSet("thresholds.cpu") {
		config.Thresholds.CPU = v.GetFloat64("thresholds.cpu")
	}
	if v.IsSet("thresholds.memory") {
		config.Thresholds.Memory = v.GetFloat64("thresholds.memory")
	}
	if v.IsSet("thresholds.disk") {
		config.Thresholds.Disk = v.GetFloat64("thresholds.disk")
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig validates configuration values
func ValidateConfig(config *Config) error {
	// Validate interval
	if config.Interval <= 0 {
		return fmt.Errorf("interval must be positive, got: %v", config.Interval)
	}
	if config.Interval > 1*time.Hour {
		return fmt.Errorf("interval too large (max 1 hour), got: %v", config.Interval)
	}

	// Validate thresholds
	if err := validateThreshold("CPU", config.Thresholds.CPU); err != nil {
		return err
	}
	if err := validateThreshold("Memory", config.Thresholds.Memory); err != nil {
		return err
	}
	if err := validateThreshold("Disk", config.Thresholds.Disk); err != nil {
		return err
	}

	return nil
}

// validateThreshold checks if a threshold value is in valid range [0, 100]
func validateThreshold(name string, value float64) error {
	if value < 0 || value > 100 {
		return fmt.Errorf("%s threshold must be between 0 and 100, got: %.2f", name, value)
	}
	return nil
}

// MergeWithFlags merges configuration with command-line flags (flags take precedence)
func MergeWithFlags(config *Config, interval time.Duration, jsonMode bool, logFile string,
	cpuThreshold, memThreshold, diskThreshold float64) (*Config, error) {

	// Create a copy to avoid modifying the original
	merged := *config

	// Override with flags if they were explicitly set
	// Note: The caller should track which flags were actually set
	if interval > 0 {
		merged.Interval = interval
	}

	merged.JSONMode = jsonMode

	if logFile != "" {
		merged.LogFile = logFile
	}

	if cpuThreshold > 0 {
		merged.Thresholds.CPU = cpuThreshold
	}
	if memThreshold > 0 {
		merged.Thresholds.Memory = memThreshold
	}
	if diskThreshold > 0 {
		merged.Thresholds.Disk = diskThreshold
	}

	// Validate merged configuration
	if err := ValidateConfig(&merged); err != nil {
		return nil, err
	}

	return &merged, nil
}
