package stats

import "github.com/sysmon/system-monitor-cli/internal/models"

// MockStatsProvider provides controllable test data for testing
type MockStatsProvider struct {
	CPUStats  *models.CPUStats
	MemStats  *models.MemoryStats
	DiskStats []models.DiskStats
	NetStats  []models.NetworkStats

	CPUError  error
	MemError  error
	DiskError error
	NetError  error
}

// NewMockStatsProvider creates a new mock stats provider with default values
func NewMockStatsProvider() *MockStatsProvider {
	return &MockStatsProvider{
		CPUStats: &models.CPUStats{
			Overall: 25.5,
			PerCore: []float64{20.0, 30.0, 25.0, 28.0},
		},
		MemStats: &models.MemoryStats{
			Total:     16 * 1024 * 1024 * 1024, // 16 GB
			Used:      8 * 1024 * 1024 * 1024,  // 8 GB
			Available: 8 * 1024 * 1024 * 1024,  // 8 GB
			Percent:   50.0,
		},
		DiskStats: []models.DiskStats{
			{
				Mountpoint: "/",
				Total:      500 * 1024 * 1024 * 1024, // 500 GB
				Used:       300 * 1024 * 1024 * 1024, // 300 GB
				Available:  200 * 1024 * 1024 * 1024, // 200 GB
				Percent:    60.0,
			},
		},
		NetStats: []models.NetworkStats{
			{
				Interface: "eth0",
				BytesSent: 1024 * 1024 * 100, // 100 MB
				BytesRecv: 1024 * 1024 * 200, // 200 MB
				SendRate:  1024 * 100,        // 100 KB/s
				RecvRate:  1024 * 200,        // 200 KB/s
			},
		},
	}
}

// GetCPUStats returns mock CPU statistics or an error
func (m *MockStatsProvider) GetCPUStats() (*models.CPUStats, error) {
	if m.CPUError != nil {
		return nil, m.CPUError
	}
	return m.CPUStats, nil
}

// GetMemoryStats returns mock memory statistics or an error
func (m *MockStatsProvider) GetMemoryStats() (*models.MemoryStats, error) {
	if m.MemError != nil {
		return nil, m.MemError
	}
	return m.MemStats, nil
}

// GetDiskStats returns mock disk statistics or an error
func (m *MockStatsProvider) GetDiskStats() ([]models.DiskStats, error) {
	if m.DiskError != nil {
		return nil, m.DiskError
	}
	return m.DiskStats, nil
}

// GetNetworkStats returns mock network statistics or an error
func (m *MockStatsProvider) GetNetworkStats() ([]models.NetworkStats, error) {
	if m.NetError != nil {
		return nil, m.NetError
	}
	return m.NetStats, nil
}
