package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"go.aimuz.me/mynt/logger"
)

// SystemStats represents the system status.
type SystemStats struct {
	CPU    CPUStats    `json:"cpu"`
	Memory MemoryStats `json:"memory"`
	Swap   SwapStats   `json:"swap"`
	Uptime uint64      `json:"uptime"`
}

// CPUStats holds CPU usage information.
type CPUStats struct {
	TotalUsage float64   `json:"total_usage"` // Total CPU usage percentage
	Cores      []float64 `json:"cores"`       // Usage per core
	Model      string    `json:"model"`
}

// MemoryStats holds memory usage information.
type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// SwapStats holds swap usage information.
type SwapStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkStats holds network interface counters.
type NetworkStats struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// ProcessInfo represents a single process.
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	Status        string  `json:"status"` // "Running", "Sleeping", etc.
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	CreateTime    int64   `json:"create_time"`
	Cmdline       string  `json:"cmdline"`
}

// processFetcher abstracts the fetching of process data to facilitate testing.
type processFetcher interface {
	Processes(ctx context.Context) ([]*process.Process, error)
}

// realProcessFetcher uses gopsutil to fetch real processes.
type realProcessFetcher struct{}

func (f *realProcessFetcher) Processes(ctx context.Context) ([]*process.Process, error) {
	return process.ProcessesWithContext(ctx)
}

// SystemMonitor handles system stats collection and process monitoring.
type SystemMonitor struct {
	mu      sync.RWMutex
	cache   map[int32]*process.Process
	fetcher processFetcher
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewSystemMonitor creates a new SystemMonitor.
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		cache:   make(map[int32]*process.Process),
		fetcher: &realProcessFetcher{},
	}
}

// Start begins the background monitoring loop.
func (m *SystemMonitor) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)
	m.wg.Add(1)
	go m.run(ctx)
}

// Stop stops the background monitoring loop.
func (m *SystemMonitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
}

func (m *SystemMonitor) run(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Initial update
	m.updateCache(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.updateCache(ctx)
		}
	}
}

func (m *SystemMonitor) updateCache(ctx context.Context) {
	procs, err := m.fetcher.Processes(ctx)
	if err != nil {
		logger.Error("failed to list processes", "error", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	newCache := make(map[int32]*process.Process)
	for _, p := range procs {
		if existing, ok := m.cache[p.Pid]; ok {
			newCache[p.Pid] = existing
		} else {
			newCache[p.Pid] = p
		}
	}
	m.cache = newCache
}

// GetSystemStats returns current system statistics.
func (m *SystemMonitor) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	// CPU
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu percent: %w", err)
	}

	// Per core
	cpuCores, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu cores percent: %w", err)
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	model := ""
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}

	// Memory
	vMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	// Swap
	sMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap memory: %w", err)
	}

	return &SystemStats{
		CPU: CPUStats{
			TotalUsage: cpuPercent[0],
			Cores:      cpuCores,
			Model:      model,
		},
		Memory: MemoryStats{
			Total:       vMem.Total,
			Available:   vMem.Available,
			Used:        vMem.Used,
			UsedPercent: vMem.UsedPercent,
		},
		Swap: SwapStats{
			Total:       sMem.Total,
			Used:        sMem.Used,
			Free:        sMem.Free,
			UsedPercent: sMem.UsedPercent,
		},
	}, nil
}

// GetProcesses returns list of all running processes.
func (m *SystemMonitor) GetProcesses(ctx context.Context) ([]ProcessInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ProcessInfo
	for _, p := range m.cache {
		name, _ := p.NameWithContext(ctx)
		username, _ := p.UsernameWithContext(ctx)
		status, _ := p.StatusWithContext(ctx)
		createTime, _ := p.CreateTimeWithContext(ctx)
		cmdline, _ := p.CmdlineWithContext(ctx)

		cpuPct, _ := p.CPUPercentWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)

		statusStr := ""
		if len(status) > 0 {
			statusStr = status[0]
		}

		result = append(result, ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			Username:      username,
			Status:        statusStr,
			CPUPercent:    cpuPct,
			MemoryPercent: memPct,
			CreateTime:    createTime,
			Cmdline:       cmdline,
		})
	}
	return result, nil
}

// KillProcess kills a process by PID.
func (m *SystemMonitor) KillProcess(ctx context.Context, pid int32) error {
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	return p.KillWithContext(ctx)
}

// GetNetworkStats returns network statistics.
func (m *SystemMonitor) GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	counters, err := net.IOCountersWithContext(ctx, false)
	if err != nil || len(counters) == 0 {
		return nil, fmt.Errorf("failed to get network counters: %w", err)
	}
	c := counters[0]
	return &NetworkStats{
		BytesSent:   c.BytesSent,
		BytesRecv:   c.BytesRecv,
		PacketsSent: c.PacketsSent,
		PacketsRecv: c.PacketsRecv,
	}, nil
}
