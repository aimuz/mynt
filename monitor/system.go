package monitor

import (
	"context"
	"fmt"

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
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var result []ProcessInfo
	for _, p := range procs {
		// Basic info
		name, _ := p.NameWithContext(ctx)
		username, _ := p.UsernameWithContext(ctx)
		status, _ := p.StatusWithContext(ctx)
		createTime, _ := p.CreateTimeWithContext(ctx)
		cmdline, _ := p.CmdlineWithContext(ctx)

		// Resources
		// Note: CPUPercent returns usage since last call, so it might be 0 on first call or instant
		// For accurate reading, we might need state, but for simple list 0 is expected initially
		cpuPct, _ := p.CPUPercentWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)

		result = append(result, ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			Username:      username,
			Status:        status[0], // usually returns []string
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
