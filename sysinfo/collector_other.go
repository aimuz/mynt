//go:build !darwin

// Package sysinfo provides system information collection.
// This file contains non-Linux implementation using gopsutil (slower but portable).
package sysinfo

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
)

// ListProcesses returns a list of running processes.
// Uses gopsutil for portability on non-Linux systems.
// ListProcesses returns a list of running processes.
func (c *Collector) ListProcesses() ([]Process, error) {
	ctx := context.Background()
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Process, 0, len(procs))
	for _, p := range procs {
		proc := Process{PID: int(p.Pid)}

		if name, err := p.NameWithContext(ctx); err == nil {
			proc.Name = name
		}

		if cmdline, err := p.CmdlineWithContext(ctx); err == nil {
			proc.Command = cmdline
		}
		if proc.Command == "" {
			proc.Command = proc.Name
		}

		if user, err := p.UsernameWithContext(ctx); err == nil {
			proc.User = user
		}

		if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
			proc.CPUPercent = cpuPct
		}

		if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil && memInfo != nil {
			proc.MemRSS = memInfo.RSS
		}

		if memPct, err := p.MemoryPercentWithContext(ctx); err == nil {
			proc.MemPercent = float64(memPct)
		}

		if status, err := p.StatusWithContext(ctx); err == nil && len(status) > 0 {
			proc.State = status[0]
		}

		if createTime, err := p.CreateTimeWithContext(ctx); err == nil {
			proc.StartTime = createTime / 1000 // Convert from ms to seconds
		}

		if threads, err := p.NumThreadsWithContext(ctx); err == nil {
			proc.Threads = int(threads)
		}

		result = append(result, proc)
	}

	return result, nil
}
