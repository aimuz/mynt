package monitor

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"go.aimuz.me/mynt/logger"
)

// SystemStats represents a snapshot of system resource usage.
type SystemStats struct {
	CPU      CPUStats                  `json:"cpu"`
	Memory   MemoryStats               `json:"memory"`
	Network  map[string]InterfaceStats `json:"network"`
	Disk     map[string]DiskStats      `json:"disk"`
	Uptime   uint64                    `json:"uptime"`
	HostInfo HostInfo                  `json:"host_info"`
}

type CPUStats struct {
	Global  float64   `json:"global"`
	PerCore []float64 `json:"per_core"`
}

type MemoryStats struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Cached    uint64 `json:"cached"`
	Free      uint64 `json:"free"`
	SwapTotal uint64 `json:"swap_total"`
	SwapUsed  uint64 `json:"swap_used"`
}

type InterfaceStats struct {
	Name      string `json:"name"`
	RxBytes   uint64 `json:"rx_bytes"` // Total bytes received
	TxBytes   uint64 `json:"tx_bytes"` // Total bytes sent
	RxRate    uint64 `json:"rx_rate"`  // Bytes per second
	TxRate    uint64 `json:"tx_rate"`  // Bytes per second
	IsUp      bool   `json:"is_up"`
	IPAddress string `json:"ip_address"`
}

type DiskStats struct {
	Name      string `json:"name"`
	ReadBytes uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadRate  uint64 `json:"read_rate"`
	WriteRate uint64 `json:"write_rate"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Kernel   string `json:"kernel"`
}

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	Username   string  `json:"username"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	MemRSS     uint64  `json:"mem_rss"`
	CmdLine    string  `json:"cmdline"`
}

// SystemDataFetcher abstracts gopsutil calls for testing.
type SystemDataFetcher interface {
	CPUPercent(interval time.Duration, percpu bool) ([]float64, error)
	VirtualMemory() (*mem.VirtualMemoryStat, error)
	SwapMemory() (*mem.SwapMemoryStat, error)
	NetIOCounters(pernic bool) ([]net.IOCountersStat, error)
	NetInterfaces() ([]net.InterfaceStat, error)
	DiskIOCounters(names ...string) (map[string]disk.IOCountersStat, error)
	Processes() ([]*process.Process, error)
	HostInfo() (*host.InfoStat, error)
}

// RealSystemFetcher uses gopsutil.
type RealSystemFetcher struct{}

func (f *RealSystemFetcher) CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	return cpu.Percent(interval, percpu)
}

func (f *RealSystemFetcher) HostInfo() (*host.InfoStat, error) {
	return host.Info()
}

func (f *RealSystemFetcher) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (f *RealSystemFetcher) SwapMemory() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}

func (f *RealSystemFetcher) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	return net.IOCounters(pernic)
}

func (f *RealSystemFetcher) NetInterfaces() ([]net.InterfaceStat, error) {
	return net.Interfaces()
}

func (f *RealSystemFetcher) DiskIOCounters(names ...string) (map[string]disk.IOCountersStat, error) {
	return disk.IOCounters(names...)
}

func (f *RealSystemFetcher) Processes() ([]*process.Process, error) {
	return process.Processes()
}

// SystemMonitor collects and caches system statistics.
type SystemMonitor struct {
	fetcher  SystemDataFetcher
	interval time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	mu          sync.RWMutex
	stats       SystemStats
	lastCollect time.Time
}

// NewSystemMonitor creates a new system monitor.
func NewSystemMonitor(interval time.Duration) *SystemMonitor {
	return &SystemMonitor{
		fetcher:  &RealSystemFetcher{},
		interval: interval,
		stats: SystemStats{
			Network: make(map[string]InterfaceStats),
			Disk:    make(map[string]DiskStats),
		},
	}
}

// Start begins the monitoring loop.
func (m *SystemMonitor) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)
	logger.Info("system monitoring started", "interval", m.interval)
	m.wg.Go(func() {
		m.run(ctx)
	})
}

// Stop halts the monitoring loop.
func (m *SystemMonitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
}

func (m *SystemMonitor) run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Initial collection
	m.collect()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collect()
		}
	}
}

func (m *SystemMonitor) collect() {
	now := time.Now()

	// CPU
	// passing 0 interval to return immediately (requires previous call for delta,
	// but gopsutil maintains internal state for this)
	// Actually gopsutil cpu.Percent needs interval > 0 to block and calculate,
	// OR it uses last call time. However, gopsutil documentation says:
	// "if interval is 0, return immediately ... using the time since last call"
	// This fits our loop perfectly.
	cpuGlobal, err := m.fetcher.CPUPercent(0, false)
	if err != nil {
		logger.Error("failed to get cpu stats", "error", err)
	}
	cpuPerCore, err := m.fetcher.CPUPercent(0, true)
	if err != nil {
		logger.Error("failed to get cpu core stats", "error", err)
	}

	// Memory
	vmem, err := m.fetcher.VirtualMemory()
	if err != nil {
		logger.Error("failed to get memory stats", "error", err)
	}
	swap, err := m.fetcher.SwapMemory()
	if err != nil {
		logger.Error("failed to get swap stats", "error", err)
	}

	// Network
	netIO, err := m.fetcher.NetIOCounters(true)
	if err != nil {
		logger.Error("failed to get net stats", "error", err)
	}
	// We also need interfaces to get names/IPs if needed, but for now we rely on IO counters names
	// To get friendly names or IPs, we might need NetInterfaces call occasionally.
	// For simplicity, we just use what IO counters gives us.

	// Disk
	diskIO, err := m.fetcher.DiskIOCounters()
	if err != nil {
		logger.Error("failed to get disk stats", "error", err)
	}

	// Host Info (only if missing)
	var hostInfo *host.InfoStat
	if m.stats.HostInfo.Hostname == "" {
		hostInfo, err = m.fetcher.HostInfo()
		if err != nil {
			logger.Error("failed to get host info", "error", err)
		}
	}

	// Calculate Rates and Update State
	m.mu.Lock()
	defer m.mu.Unlock()

	timeDiff := now.Sub(m.lastCollect).Seconds()
	if timeDiff <= 0 {
		timeDiff = 1.0 // Prevent division by zero on first run
	}

	// Update CPU
	if len(cpuGlobal) > 0 {
		m.stats.CPU.Global = cpuGlobal[0]
	}
	m.stats.CPU.PerCore = cpuPerCore

	// Update Memory
	if vmem != nil {
		m.stats.Memory.Total = vmem.Total
		m.stats.Memory.Used = vmem.Used
		m.stats.Memory.Cached = vmem.Cached
		m.stats.Memory.Free = vmem.Free
	}
	if swap != nil {
		m.stats.Memory.SwapTotal = swap.Total
		m.stats.Memory.SwapUsed = swap.Used
	}

	// Update Network
	newNetMap := make(map[string]InterfaceStats)
	for _, io := range netIO {
		prev, exists := m.stats.Network[io.Name]
		stat := InterfaceStats{
			Name:    io.Name,
			RxBytes: io.BytesRecv,
			TxBytes: io.BytesSent,
		}
		if exists && m.lastCollect.IsZero() == false {
			stat.RxRate = uint64(float64(io.BytesRecv-prev.RxBytes) / timeDiff)
			stat.TxRate = uint64(float64(io.BytesSent-prev.TxBytes) / timeDiff)
		}
		newNetMap[io.Name] = stat
	}
	m.stats.Network = newNetMap

	// Update Disk
	newDiskMap := make(map[string]DiskStats)
	for name, io := range diskIO {
		prev, exists := m.stats.Disk[name]
		stat := DiskStats{
			Name:       name,
			ReadBytes:  io.ReadBytes,
			WriteBytes: io.WriteBytes,
		}
		if exists && m.lastCollect.IsZero() == false {
			stat.ReadRate = uint64(float64(io.ReadBytes-prev.ReadBytes) / timeDiff)
			stat.WriteRate = uint64(float64(io.WriteBytes-prev.WriteBytes) / timeDiff)
		}
		newDiskMap[name] = stat
	}
	m.stats.Disk = newDiskMap

	if hostInfo != nil {
		m.stats.HostInfo = HostInfo{
			Hostname: hostInfo.Hostname,
			OS:       hostInfo.OS,
			Platform: hostInfo.Platform,
			Kernel:   hostInfo.KernelVersion,
		}
		m.stats.Uptime = hostInfo.Uptime
	} else if m.stats.HostInfo.Hostname != "" {
		// Just update uptime
		// We could fetch uptime separately but gopsutil HostInfo provides it.
		// For simplicity, we can fetch HostInfo every time or just calculate locally.
		// Re-fetching HostInfo is cheap enough (reads /proc/uptime).
		h, err := m.fetcher.HostInfo()
		if err == nil {
			m.stats.Uptime = h.Uptime
		}
	}

	m.lastCollect = now
}

// GetStats returns the latest collected stats.
func (m *SystemMonitor) GetStats() SystemStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	// Maps need deep copy
	stats := m.stats
	stats.Network = make(map[string]InterfaceStats, len(m.stats.Network))
	for k, v := range m.stats.Network {
		stats.Network[k] = v
	}
	stats.Disk = make(map[string]DiskStats, len(m.stats.Disk))
	for k, v := range m.stats.Disk {
		stats.Disk[k] = v
	}
	// PerCore slice copy
	stats.CPU.PerCore = make([]float64, len(m.stats.CPU.PerCore))
	copy(stats.CPU.PerCore, m.stats.CPU.PerCore)

	return stats
}

// GetProcesses returns the current list of processes.
// This is done on demand as it is heavy.
func (m *SystemMonitor) GetProcesses() ([]ProcessInfo, error) {
	procs, err := m.fetcher.Processes()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	var infos []ProcessInfo
	for _, p := range procs {
		// We handle errors for individual process attributes gracefully by logging debug
		// or using sensible defaults, but strict rsc style prefers checking err.
		// However, process attributes often fail due to permissions or race conditions (process died).
		// We accumulate valid data where possible.

		name, err := p.Name()
		if err != nil {
			// If we can't get name, process might be dead or inaccessible. Skip or continue.
			// Usually skipping is safer for UI.
			continue
		}

		username, err := p.Username()
		if err != nil {
			username = ""
		}

		statusSlice, err := p.Status()
		status := "unknown"
		if err == nil && len(statusSlice) > 0 {
			status = statusSlice[0]
		}

		cpuP, err := p.CPUPercent()
		if err != nil {
			cpuP = 0
		}

		memP, err := p.MemoryPercent()
		if err != nil {
			memP = 0
		}

		memInfo, err := p.MemoryInfo()
		var rss uint64
		if err == nil && memInfo != nil {
			rss = memInfo.RSS
		}

		cmd, err := p.Cmdline()
		if err != nil {
			cmd = ""
		}

		info := ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			Username:   username,
			Status:     status,
			CPUPercent: cpuP,
			MemPercent: memP,
			CmdLine:    cmd,
			MemRSS:     rss,
		}

		infos = append(infos, info)
	}

	// Sort by CPU usage desc by default
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].CPUPercent > infos[j].CPUPercent
	})

	return infos, nil
}

// KillProcess terminates a process.
func (m *SystemMonitor) KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	if err := p.Kill(); err != nil {
		return fmt.Errorf("kill process: %w", err)
	}
	return nil
}
