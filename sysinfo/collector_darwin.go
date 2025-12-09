//go:build darwin

// Package sysinfo provides system information collection.
// This file contains macOS-specific implementation using:
// - sysctl for batch PID/basic info retrieval (ONE syscall for all processes)
// - CGO + libproc for Memory/Threads/CPU time (ONE call per process)
// - Delta-based CPU% calculation using lastCPU snapshots
package sysinfo

/*
#include <libproc.h>
#include <sys/proc_info.h>
#include <mach/mach_time.h>

// Get Mach timebase info for converting Mach absolute time to nanoseconds
static double getMachTimebaseNsPerTick() {
    mach_timebase_info_data_t info;
    mach_timebase_info(&info);
    return (double)info.numer / (double)info.denom;
}

// Get process status using proc_taskallinfo
// On macOS, most processes in sleep/wait show pbi_status=2 (SRUN) but have
// different thread states. We check pti_numrunning to distinguish.
// Returns: 1=SIDL, 2=SRUN, 3=SSLEEP, 4=SSTOP, 5=SZOMB
static int getProcessStatus(int pid) {
    struct proc_taskallinfo info;
    int ret = proc_pidinfo(pid, PROC_PIDTASKALLINFO, 0, &info, sizeof(info));
    if (ret == sizeof(info)) {
        // pbi_status values: SIDL=1, SRUN=2, SSLEEP=3, SSTOP=4, SZOMB=5
        // On macOS, sleeping processes often show SRUN, so we check:
        // - If pbi_status is SRUN (2) and pti_numrunning == 0, it's sleeping
        if (info.pbsd.pbi_status == 2 && info.ptinfo.pti_numrunning == 0) {
            return 3; // SSLEEP
        }
        return info.pbsd.pbi_status;
    }
    return 2; // default to SRUN if we can't get status
}
*/
import "C"

import (
	"os/user"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// machTimebaseNsPerTick is the conversion factor from Mach absolute time to nanoseconds.
// Cached at init time for performance.
var machTimebaseNsPerTick = float64(C.getMachTimebaseNsPerTick())

// ListProcesses returns a list of running processes.
func (c *Collector) ListProcesses() ([]Process, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Step 1: Get all processes using sysctl (ONE syscall for all PIDs)
	kprocs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	elapsed := now.Sub(c.lastCPUTime).Seconds()

	// Get total physical memory using sysctl (same approach as CPU)
	totalMem, _ := unix.SysctlUint64("hw.memsize")

	// Build new CPU snapshot map
	newCPU := make(map[int]cpuSnapshot, len(kprocs))

	result := make([]Process, 0, len(kprocs))
	for i := range kprocs {
		kp := &kprocs[i]
		pid := int(kp.Proc.P_pid)
		if pid <= 0 {
			continue
		}

		// Basic info from sysctl (fast, already in memory)
		name := cstring(kp.Proc.P_comm[:])
		uid := kp.Eproc.Ucred.Uid

		// StartTime in milliseconds
		createtime := kp.Proc.P_starttime.Sec*1000 + int64(kp.Proc.P_starttime.Usec)/1000

		proc := Process{
			PID:       pid,
			Name:      name,
			User:      lookupUsername(int(uid)),
			State:     processState(int8(C.getProcessStatus(C.int(pid)))),
			StartTime: createtime,
		}

		proc.Command = name
		// Get command line using sysctl
		if args := getProcArgs(pid); args != "" {
			proc.Command = args
		}

		// Step 2: Get Memory/Threads/CPU using proc_pidinfo (ONE call per process)
		var taskInfo C.struct_proc_taskinfo
		ret := C.proc_pidinfo(C.int(pid), C.PROC_PIDTASKINFO, 0,
			unsafe.Pointer(&taskInfo), C.int(C.PROC_PIDTASKINFO_SIZE))

		if ret == C.PROC_PIDTASKINFO_SIZE {
			proc.MemRSS = uint64(taskInfo.pti_resident_size)
			proc.Threads = int(taskInfo.pti_threadnum)

			// Calculate MemPercent using sysctl-retrieved total memory
			if totalMem > 0 {
				proc.MemPercent = float64(proc.MemRSS) / float64(totalMem) * 100.0
			}

			// CPU time in seconds:
			// pti_total_* are in Mach absolute time units, convert using timebase
			machTime := float64(taskInfo.pti_total_system + taskInfo.pti_total_user)
			cpuTimeSec := (machTime * machTimebaseNsPerTick) / 1e9

			// Store snapshot for next calculation
			newCPU[pid] = cpuSnapshot{cpuTime: cpuTimeSec, at: now}

			// Calculate CPU percentage using delta from previous sample
			if prev, ok := c.lastCPU[pid]; ok && elapsed > 0 {
				deltaCPU := cpuTimeSec - prev.cpuTime
				if deltaCPU >= 0 {
					proc.CPUPercent = (deltaCPU / elapsed) * 100.0
				}
			}
		}

		result = append(result, proc)
	}

	// Update snapshots for next call
	c.lastCPU = newCPU
	c.lastCPUTime = now

	return result, nil
}

// getProcArgs retrieves command line arguments using sysctl
func getProcArgs(pid int) string {
	args, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil || len(args) < 4 {
		return ""
	}

	data := args[4:]
	nullIdx := strings.IndexByte(string(data), 0)
	if nullIdx == -1 {
		return ""
	}

	start := nullIdx + 1
	for start < len(data) && data[start] == 0 {
		start++
	}

	if start >= len(data) {
		return ""
	}

	var argList []string
	for start < len(data) {
		end := start
		for end < len(data) && data[end] != 0 {
			end++
		}
		if end == start {
			break
		}
		argList = append(argList, string(data[start:end]))
		start = end + 1
	}

	return strings.Join(argList, " ")
}

// cstring converts a null-terminated byte array to string
func cstring(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// processState converts macOS process stat to state string
func processState(stat int8) string {
	switch stat {
	case 2:
		return "R"
	case 3:
		return "S"
	case 4:
		return "T"
	case 5:
		return "Z"
	default:
		return "S"
	}
}

// lookupUsername returns username for a UID
func lookupUsername(uid int) string {
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return strconv.Itoa(uid)
	}
	return u.Username
}
