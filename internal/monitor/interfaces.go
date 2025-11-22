package monitor

import "context"

// Monitor defines the interface for the system monitor orchestrator
type Monitor interface {
	// Start begins monitoring and blocks until context is cancelled
	Start(ctx context.Context) error

	// Stop performs cleanup and stops monitoring
	Stop() error
}
