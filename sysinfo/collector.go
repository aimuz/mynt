// Package sysinfo provides system information collection using gopsutil.
package sysinfo

import (
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// cpuSnapshot stores CPU time for rate calculation.
type cpuSnapshot struct {
	cpuTime float64
	at      time.Time
}

// Collector gathers system statistics using gopsutil.
// It maintains snapshots to calculate rates.
type Collector struct {
	mu       sync.Mutex
	lastNet  map[string]netSnapshot
	lastDisk map[string]diskSnapshot
	lastTime time.Time

	// Process CPU snapshots for percentage calculation
	lastCPU     map[int]cpuSnapshot
	lastCPUTime time.Time
	uidCache    map[int]string
	readBuf     [4096]byte
}

type netSnapshot struct {
	bytesIn  uint64
	bytesOut uint64
}

type diskSnapshot struct {
	readBytes  uint64
	writeBytes uint64
}

// NewCollector creates a new system info collector.
func NewCollector() *Collector {
	return &Collector{
		lastNet:  make(map[string]netSnapshot),
		lastDisk: make(map[string]diskSnapshot),
		lastCPU:  make(map[int]cpuSnapshot),
		uidCache: make(map[int]string),
	}
}

// Collect gathers current system statistics.
func (c *Collector) Collect() (*Stats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastTime).Seconds()
	skipSpeeds := c.lastTime.IsZero() || elapsed < 0.1
	if skipSpeeds {
		elapsed = 1.0 // Prevent division by zero
	}

	stats := &Stats{
		CPU:     CPUStats{},
		Memory:  MemStats{},
		Network: []NetStats{},
		DiskIO:  []DiskIO{},
	}

	// CPU stats
	if percents, err := cpu.Percent(0, true); err == nil {
		stats.CPU.Cores = percents
		stats.CPU.CoreCount = len(percents)
		var total float64
		for _, p := range percents {
			total += p
		}
		if len(percents) > 0 {
			stats.CPU.Total = total / float64(len(percents))
		}
	}

	// CPU frequency
	if infos, err := cpu.Info(); err == nil && len(infos) > 0 {
		stats.CPU.Frequency = infos[0].Mhz
	}

	// Memory stats
	if vmem, err := mem.VirtualMemory(); err == nil {
		stats.Memory.Total = vmem.Total
		stats.Memory.Used = vmem.Used
		stats.Memory.Available = vmem.Available
		stats.Memory.Cached = vmem.Cached
		stats.Memory.Buffers = vmem.Buffers
		stats.Memory.Percent = vmem.UsedPercent
	}

	// Swap
	if swap, err := mem.SwapMemory(); err == nil {
		stats.Memory.SwapTotal = swap.Total
		stats.Memory.SwapUsed = swap.Used
	}

	// Uptime
	if uptime, err := host.Uptime(); err == nil {
		stats.Uptime = uptime
	}

	// Network stats
	if counters, err := net.IOCounters(true); err == nil {
		newNet := make(map[string]netSnapshot)
		for _, ioc := range counters {
			if ioc.Name == "lo" {
				continue // Skip loopback
			}

			ns := NetStats{
				Name:     ioc.Name,
				BytesIn:  ioc.BytesRecv,
				BytesOut: ioc.BytesSent,
				IsUp:     ioc.BytesRecv > 0 || ioc.BytesSent > 0,
			}

			// Calculate speed if we have previous data
			if prev, ok := c.lastNet[ioc.Name]; ok && elapsed > 0 {
				ns.SpeedIn = float64(ioc.BytesRecv-prev.bytesIn) / elapsed
				ns.SpeedOut = float64(ioc.BytesSent-prev.bytesOut) / elapsed
			}

			newNet[ioc.Name] = netSnapshot{
				bytesIn:  ioc.BytesRecv,
				bytesOut: ioc.BytesSent,
			}

			stats.Network = append(stats.Network, ns)
		}
		c.lastNet = newNet
	}

	// Disk I/O stats
	if counters, err := disk.IOCounters(); err == nil {
		newDisk := make(map[string]diskSnapshot)
		for name, ioc := range counters {
			dio := DiskIO{
				Device:     name,
				ReadBytes:  ioc.ReadBytes,
				WriteBytes: ioc.WriteBytes,
			}

			// Calculate speed if we have previous data
			if prev, ok := c.lastDisk[name]; ok && elapsed > 0 {
				dio.ReadSpeed = float64(ioc.ReadBytes-prev.readBytes) / elapsed
				dio.WriteSpeed = float64(ioc.WriteBytes-prev.writeBytes) / elapsed
			}

			newDisk[name] = diskSnapshot{
				readBytes:  ioc.ReadBytes,
				writeBytes: ioc.WriteBytes,
			}

			stats.DiskIO = append(stats.DiskIO, dio)
		}
		c.lastDisk = newDisk
	}

	c.lastTime = now
	return stats, nil
}

// KillProcess sends a signal to a process.
func (c *Collector) KillProcess(pid int, signal syscall.Signal) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	return p.SendSignal(signal)
}
