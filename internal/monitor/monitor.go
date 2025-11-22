package monitor

import (
	"context"
	"sync"

	"github.com/sysmon/system-monitor-cli/internal/collector"
	"github.com/sysmon/system-monitor-cli/internal/config"
	"github.com/sysmon/system-monitor-cli/internal/logger"
	"github.com/sysmon/system-monitor-cli/internal/models"
	"github.com/sysmon/system-monitor-cli/internal/render"
)

// SystemMonitor orchestrates the monitoring application lifecycle
type SystemMonitor struct {
	config    *config.Config
	collector collector.MetricsCollector
	renderer  render.Renderer
	logger    logger.Logger
	wg        sync.WaitGroup
}

// NewSystemMonitor creates a new system monitor instance
func NewSystemMonitor(
	cfg *config.Config,
	collector collector.MetricsCollector,
	renderer render.Renderer,
	logger logger.Logger,
) *SystemMonitor {
	return &SystemMonitor{
		config:    cfg,
		collector: collector,
		renderer:  renderer,
		logger:    logger,
	}
}

// Start begins monitoring and blocks until context is cancelled
func (m *SystemMonitor) Start(ctx context.Context) error {
	metricsChan := make(chan *models.Metrics, 1)

	// Start collector goroutine
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.collector.Start(ctx, m.config.Interval, metricsChan)
	}()

	// Main loop - receive and render metrics
	for {
		select {
		case <-ctx.Done():
			// Wait for collector to finish
			m.wg.Wait()
			return ctx.Err()
		case metrics, ok := <-metricsChan:
			if !ok {
				// Channel closed, collector stopped
				return nil
			}

			// Render metrics
			if err := m.renderer.Render(metrics); err != nil {
				// Log error but continue
				if m.logger != nil {
					m.logger.LogError(err)
				}
			}

			// Log metrics if logger is configured
			if m.logger != nil {
				if err := m.logger.LogMetrics(metrics); err != nil {
					// Errors are already logged to stderr by the logger
					continue
				}
			}
		}
	}
}

// Stop performs cleanup and stops monitoring
func (m *SystemMonitor) Stop() error {
	// Clear renderer
	if err := m.renderer.Clear(); err != nil {
		// Ignore clear errors
	}

	// Close renderer
	if err := m.renderer.Close(); err != nil {
		return err
	}

	// Close logger if present
	if m.logger != nil {
		if err := m.logger.Close(); err != nil {
			return err
		}
	}

	return nil
}
