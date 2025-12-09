package monitor

import (
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

// MockFetcher implements SystemDataFetcher for testing.
type MockFetcher struct {
	cpuPercent   []float64
	vMem         *mem.VirtualMemoryStat
	swap         *mem.SwapMemoryStat
	netIO        []net.IOCountersStat
	diskIO       map[string]disk.IOCountersStat
	processes    []*process.Process
}

func (m *MockFetcher) HostInfo() (*host.InfoStat, error) {
	return &host.InfoStat{
		Hostname: "test-host",
		OS:       "linux",
		Platform: "mynt",
		Uptime:   12345,
	}, nil
}

func (m *MockFetcher) CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	if percpu {
		return []float64{10.0, 20.0, 30.0, 40.0}, nil
	}
	return []float64{25.0}, nil
}

func (m *MockFetcher) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return m.vMem, nil
}

func (m *MockFetcher) SwapMemory() (*mem.SwapMemoryStat, error) {
	return m.swap, nil
}

func (m *MockFetcher) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	return m.netIO, nil
}

func (m *MockFetcher) NetInterfaces() ([]net.InterfaceStat, error) {
	return nil, nil
}

func (m *MockFetcher) DiskIOCounters(names ...string) (map[string]disk.IOCountersStat, error) {
	return m.diskIO, nil
}

func (m *MockFetcher) Processes() ([]*process.Process, error) {
	return m.processes, nil
}

func TestSystemMonitor_Collect(t *testing.T) {
	tests := []struct {
		name     string
		pass1    func(*MockFetcher) // Setup for first pass
		pass2    func(*MockFetcher) // Setup for second pass
		validate func(*testing.T, SystemStats)
	}{
		{
			name: "Basic Stats Collection",
			pass1: func(m *MockFetcher) {
				m.vMem = &mem.VirtualMemoryStat{Total: 1000, Used: 500}
				m.swap = &mem.SwapMemoryStat{Total: 2000, Used: 100}
				m.netIO = []net.IOCountersStat{{Name: "eth0", BytesRecv: 100, BytesSent: 200}}
				m.diskIO = map[string]disk.IOCountersStat{"sda": {ReadBytes: 1000, WriteBytes: 2000}}
			},
			pass2: func(m *MockFetcher) {
				// Incremented counters to test rates
				m.netIO = []net.IOCountersStat{{Name: "eth0", BytesRecv: 200, BytesSent: 400}} // +100, +200
				m.diskIO = map[string]disk.IOCountersStat{"sda": {ReadBytes: 1500, WriteBytes: 3000}} // +500, +1000
			},
			validate: func(t *testing.T, stats SystemStats) {
				assert.Equal(t, uint64(1000), stats.Memory.Total)
				assert.Equal(t, uint64(500), stats.Memory.Used)
				assert.Equal(t, 25.0, stats.CPU.Global)
				assert.Len(t, stats.CPU.PerCore, 4)

				// Rate calculation check (assuming approx 1s interval)
				// Note: In test we call collect() manually without waiting 1s, so timeDiff might be small.
				// However, Monitor uses time.Now() diffs.
				// We need to simulate time passing for rates to be meaningful numbers.
				// For this test, we verify that counters updated and map entries exist.

				assert.Contains(t, stats.Network, "eth0")
				assert.Contains(t, stats.Disk, "sda")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &MockFetcher{}
			mon := NewSystemMonitor(100 * time.Millisecond)
			mon.fetcher = fetcher // Inject mock

			// Pass 1
			tt.pass1(fetcher)
			mon.collect()

			// Sleep to simulate interval for rate calculation
			time.Sleep(100 * time.Millisecond)

			// Pass 2
			if tt.pass2 != nil {
				tt.pass2(fetcher)
			}
			mon.collect()

			// Validate
			stats := mon.GetStats()
			tt.validate(t, stats)
		})
	}
}

func TestSystemMonitor_Rates(t *testing.T) {
	// Specialized test for rate calculation
	fetcher := &MockFetcher{}
	mon := NewSystemMonitor(time.Second)
	mon.fetcher = fetcher

	// T0
	fetcher.netIO = []net.IOCountersStat{{Name: "eth0", BytesRecv: 1000}}
	mon.collect()

	// Simulate 1 second passing
	time.Sleep(1 * time.Second)

	// T1 (1000 bytes delta)
	fetcher.netIO = []net.IOCountersStat{{Name: "eth0", BytesRecv: 2000}}
	mon.collect()

	stats := mon.GetStats()
	rate := stats.Network["eth0"].RxRate

	// Should be around 1000 bytes/sec
	// Allow small margin of error due to time.Sleep precision
	assert.InDelta(t, 1000.0, float64(rate), 100.0)
}
