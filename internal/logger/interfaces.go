package logger

import "github.com/sysmon/system-monitor-cli/internal/models"

// Logger defines the interface for logging metrics and errors
type Logger interface {
	// LogMetrics writes metrics to the log destination
	LogMetrics(metrics *models.Metrics) error

	// LogError writes an error to the log destination
	LogError(err error) error

	// Close flushes and closes the logger
	Close() error
}
