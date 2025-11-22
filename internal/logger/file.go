package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// FileLogger implements Logger for file-based logging
type FileLogger struct {
	file    *os.File
	encoder *json.Encoder
}

// NewFileLogger creates a new file logger
func NewFileLogger(path string) (*FileLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileLogger{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// LogMetrics writes metrics to the log file with timestamp
func (l *FileLogger) LogMetrics(metrics *models.Metrics) error {
	// Create log entry with ISO 8601 timestamp
	entry := struct {
		Timestamp string          `json:"timestamp"`
		Metrics   *models.Metrics `json:"metrics"`
	}{
		Timestamp: metrics.Timestamp.Format(time.RFC3339),
		Metrics:   metrics,
	}

	if err := l.encoder.Encode(entry); err != nil {
		// Log to stderr but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to write to log file: %v\n", err)
		return err
	}

	return nil
}

// LogError writes an error to the log file
func (l *FileLogger) LogError(err error) error {
	entry := struct {
		Timestamp string `json:"timestamp"`
		Error     string `json:"error"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Error:     err.Error(),
	}

	if encodeErr := l.encoder.Encode(entry); encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write error to log file: %v\n", encodeErr)
		return encodeErr
	}

	return nil
}

// Close flushes and closes the log file
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
