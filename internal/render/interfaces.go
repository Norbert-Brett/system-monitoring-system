package render

import "github.com/sysmon/system-monitor-cli/internal/models"

// Renderer defines the interface for rendering metrics
type Renderer interface {
	// Render formats and displays the given metrics
	Render(metrics *models.Metrics) error

	// Clear clears the display
	Clear() error

	// Close performs cleanup and releases resources
	Close() error
}
