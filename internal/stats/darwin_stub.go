//go:build !darwin && !linux

package stats

import "github.com/sysmon/system-monitor-cli/internal/collector"

func newPlatformProvider() collector.SystemStatsProvider {
	return nil
}
