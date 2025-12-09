package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.aimuz.me/mynt/logger"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// SystemStats represents the system status
type SystemStats struct {
	CPU    CPUStats    `json:"cpu"`
	Memory MemoryStats `json:"memory"`
	Swap   SwapStats   `json:"swap"`
	Uptime uint64      `json:"uptime"`
}

type CPUStats struct {
	TotalUsage float64   `json:"total_usage"` // Total CPU usage percentage
	Cores      []float64 `json:"cores"`       // Usage per core
	Model      string    `json:"model"`
}

type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type SwapStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// ProcessInfo represents a single process
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

// Global process cache to enable CPU percentage calculation
var (
	procCache     = make(map[int32]*process.Process)
	procCacheLock sync.RWMutex
	lastUpdate    time.Time
)

// InitProcessMonitor starts the background process monitor
func InitProcessMonitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updateProcessCache(ctx)
			}
		}
	}()
}

func updateProcessCache(ctx context.Context) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		logger.Error("failed to list processes", "error", err)
		return
	}

	procCacheLock.Lock()
	defer procCacheLock.Unlock()

	// Create a new map to track current processes
	newCache := make(map[int32]*process.Process)

	for _, p := range procs {
		// If we already have this process, keep the existing object
		// This preserves the internal state needed for CPU calculation
		if existing, ok := procCache[p.Pid]; ok {
			newCache[p.Pid] = existing
		} else {
			newCache[p.Pid] = p
		}
	}

	procCache = newCache
	lastUpdate = time.Now()
}

// GetSystemStats returns current system statistics
func GetSystemStats(ctx context.Context) (*SystemStats, error) {
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

// GetProcesses returns list of all running processes
func GetProcesses(ctx context.Context) ([]ProcessInfo, error) {
	procCacheLock.RLock()
	defer procCacheLock.RUnlock()

	// If cache is empty (first run), try to populate it immediately
	if len(procCache) == 0 {
		procCacheLock.RUnlock()
		updateProcessCache(ctx)
		procCacheLock.RLock()
	}

	var result []ProcessInfo
	for _, p := range procCache {
		// Basic info
		name, _ := p.NameWithContext(ctx)
		username, _ := p.UsernameWithContext(ctx)
		status, _ := p.StatusWithContext(ctx)
		createTime, _ := p.CreateTimeWithContext(ctx)
		cmdline, _ := p.CmdlineWithContext(ctx)

		// Resources
		// CPUPercent uses the internal state of the process object
		cpuPct, _ := p.CPUPercentWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)

		// Handle status slice safety
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

// KillProcess kills a process by PID
func KillProcess(ctx context.Context, pid int32) error {
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	return p.KillWithContext(ctx)
}

// NetworkStats
type NetworkStats struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

func GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	counters, err := net.IOCountersWithContext(ctx, false)
	if err != nil || len(counters) == 0 {
		return nil, err
	}
	c := counters[0]
	return &NetworkStats{
		BytesSent:   c.BytesSent,
		BytesRecv:   c.BytesRecv,
		PacketsSent: c.PacketsSent,
		PacketsRecv: c.PacketsRecv,
	}, nil
}
