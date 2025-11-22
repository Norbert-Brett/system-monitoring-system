package stats

import (
	"fmt"
	"runtime"

	"github.com/sysmon/system-monitor-cli/internal/collector"
)

// NewProvider creates the appropriate SystemStatsProvider for the current OS
func NewProvider() (collector.SystemStatsProvider, error) {
	provider := newPlatformProvider()
	if provider == nil {
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return provider, nil
}
