//go:build linux

// Package sysinfo provides system information collection.
// This file contains Linux-specific implementation using direct /proc parsing
// for maximum performance. Avoids third-party libraries like procfs or gopsutil.
package sysinfo

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// clkTck is the system clock ticks per second.
// On Linux, this is always 100 (USER_HZ) for the values in /proc.
// See: https://man7.org/linux/man-pages/man5/proc.5.html
const clkTck = 100.0

// bootTime is the system boot time in seconds since epoch.
// Cached at init time from /proc/stat.
var bootTime = getBootTime()

// getBootTime reads boot time from /proc/stat.
func getBootTime() int64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}
	// Scan lines without allocating a slice
	for len(data) > 0 {
		line := data
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			line = data[:i]
			data = data[i+1:]
		} else {
			data = nil
		}
		if bytes.HasPrefix(line, []byte("btime ")) {
			if v, err := strconv.ParseInt(string(line[6:]), 10, 64); err == nil {
				return v
			}
		}
	}
	return 0
}

// memTotal caches total system memory in bytes.
var memTotal uint64

func init() {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	// Scan lines without allocating a slice
	for len(data) > 0 {
		line := data
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			line = data[:i]
			data = data[i+1:]
		} else {
			data = nil
		}
		if bytes.HasPrefix(line, []byte("MemTotal:")) {
			// Format: "MemTotal:       16384000 kB"
			// Skip "MemTotal:" and whitespace, parse the number
			if v := parseFirstNumber(line[9:]); v > 0 {
				memTotal = v * 1024 // Convert kB to bytes
			}
			break
		}
	}
}

// ListProcesses returns a list of running processes.
// Uses direct /proc parsing for maximum performance on Linux.
func (c *Collector) ListProcesses() ([]Process, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastCPUTime).Seconds()

	size := max(256, len(c.lastCPU)+32)
	// Build new CPU snapshot map
	newCPU := make(map[int]cpuSnapshot, size)
	result := make([]Process, 0, size)
	err := walkPids(func(pid int) bool {
		proc, cpuTime, ok := c.parseProcStat(pid)
		if !ok {
			return true
		}
		proc.PID = pid

		// Parse /proc/[pid]/status for UID and VmRSS
		c.parseStatus(pid, &proc)

		// Parse /proc/[pid]/cmdline for full command
		if cmdline := c.readCmdline(pid); cmdline != "" {
			proc.Command = cmdline
		} else {
			proc.Command = proc.Name
		}

		// Calculate memory percentage
		if memTotal > 0 {
			proc.MemPercent = float64(proc.MemRSS) / float64(memTotal) * 100.0
		}

		// Store CPU snapshot for next calculation
		newCPU[proc.PID] = cpuSnapshot{cpuTime: cpuTime, at: now}

		// Calculate CPU percentage using delta from previous sample
		if prev, ok := c.lastCPU[proc.PID]; ok && elapsed > 0 {
			deltaCPU := cpuTime - prev.cpuTime
			if deltaCPU >= 0 {
				proc.CPUPercent = (deltaCPU / elapsed) * 100.0
			}
		}

		result = append(result, proc)
		return true
	})
	if err != nil {
		return nil, err
	}

	// Update snapshots for next call
	c.lastCPU = newCPU
	c.lastCPUTime = now

	return result, nil
}

// parseProcStat parses /proc/[pid]/stat and returns Process info and CPU time.
// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags minflt cminflt
//
//	majflt cmajflt utime stime cutime cstime priority nice num_threads
//	itrealvalue starttime vsize rss ...
func (c *Collector) parseProcStat(pid int) (proc Process, cpuTime float64, ok bool) {
	path := make([]byte, 0, 32)
	path = append(path, "/proc/"...)
	path = strconv.AppendInt(path, int64(pid), 10)
	path = append(path, "/stat"...)
	data, err := c.read(path)
	if err != nil {
		return proc, 0, false
	}

	commEnd := bytes.LastIndexByte(data, ')')
	if commEnd == -1 {
		return proc, 0, false
	}
	var n int
	var utime, stime float64
	// Field indices (0-based, after comm):
	// 2902 (YDService) S 1 2902 2532 0 -1 4194560 1182588 229385 334 5 274393 288991 9948 3013 20 0 23 0 3789 1002594304 28244 536870912 4194304 30599985 140731828525200 0 0 0 0 69635 20200 0 0 0 17 0 0 0 0 0 0 32697144 43493211 650186752 140731828526640 140731828526683 140731828526683 140731828527053 0
	// 0: state, 11: utime, 12: stime, 17: num_threads, 19: starttime
	for field := range bytes.FieldsSeq(data[commEnd+2:]) {
		switch n {
		case 0:
			proc.State = string(field)
		case 11:
			utime, _ = strconv.ParseFloat(string(field), 64)
		case 12:
			stime, _ = strconv.ParseFloat(string(field), 64)
		case 17:
			proc.Threads, _ = strconv.Atoi(string(field))
		case 19:
			starttime, _ := strconv.ParseInt(string(field), 10, 64)
			// Convert starttime from clock ticks since boot to Unix timestamp in milliseconds
			proc.StartTime = (bootTime + starttime/int64(clkTck)) * 1000
		}
		n++
		if n > 19 {
			break
		}
	}

	cpuTime = (utime + stime) / clkTck // Convert to seconds
	return proc, cpuTime, true
}

var errEAGAIN error = syscall.EAGAIN
var errEINVAL error = syscall.EINVAL
var errENOENT error = syscall.ENOENT

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.EAGAIN:
		return errEAGAIN
	case syscall.EINVAL:
		return errEINVAL
	case syscall.ENOENT:
		return errENOENT
	}
	return e
}

func open(path []byte) (fd int, err error) {
	var buf [128]byte
	n := copy(buf[:], path)
	buf[n] = 0
	dirfd := int(unix.AT_FDCWD)
	r1, _, errno := syscall.Syscall(
		syscall.SYS_OPENAT,
		uintptr(dirfd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unix.O_RDONLY|unix.O_LARGEFILE),
	)
	if errno != 0 {
		err = errnoErr(errno)
	}
	return int(r1), err
}

// This is not thread safe.
// TODO: dynamically resize the buffer.
func (c *Collector) read(path []byte) ([]byte, error) {
	fd, err := open(path)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	n, err := unix.Read(fd, c.readBuf[:])
	if err != nil {
		return nil, err
	}
	return c.readBuf[:n], nil
}

// parseStatus parses /proc/[pid]/status for UID and VmRSS.
func (c *Collector) parseStatus(pid int, proc *Process) {
	path := make([]byte, 0, 32)
	path = append(path, "/proc/"...)
	path = strconv.AppendInt(path, int64(pid), 10)
	path = append(path, "/status"...)
	data, err := c.read(path)
	if err != nil {
		return
	}

	// Scan lines without allocating a slice
	var foundUID, foundRSS bool
	for line := range bytes.Lines(data) {
		if bytes.HasPrefix(line, []byte("Name:")) {
			proc.Name = string(line[6:])
		}
		if !foundUID && bytes.HasPrefix(line, []byte("Uid:")) {
			// Format: "Uid:\t1000\t1000\t1000\t1000"
			// Skip "Uid:" and parse first number
			if uid := parseFirstNumber(line[4:]); uid > 0 || bytes.Contains(line[4:], []byte("0")) {
				proc.User = c.lookupUsername(int(uid))
			}
			foundUID = true
		} else if !foundRSS && bytes.HasPrefix(line, []byte("VmRSS:")) {
			// Format: "VmRSS:\t   12345 kB"
			if v := parseFirstNumber(line[6:]); v > 0 {
				proc.MemRSS = v * 1024 // Convert kB to bytes
			}
			foundRSS = true
		}
		if foundUID && foundRSS {
			break
		}
	}
}

// parseFirstNumber extracts the first number from a byte slice, skipping leading whitespace.
func parseFirstNumber(b []byte) uint64 {
	// Skip leading whitespace
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t') {
		i++
	}
	if i >= len(b) {
		return 0
	}
	// Find end of number
	j := i
	for j < len(b) && b[j] >= '0' && b[j] <= '9' {
		j++
	}
	if j == i {
		return 0
	}
	v, _ := atoi(b[i:j])
	return uint64(v)
}

// readCmdline reads /proc/[pid]/cmdline and returns the command line string.
func (c *Collector) readCmdline(pid int) string {
	path := make([]byte, 0, 32)
	path = append(path, "/proc/"...)
	path = strconv.AppendInt(path, int64(pid), 10)
	path = append(path, "/cmdline"...)
	data, err := c.read(path)
	if err != nil || len(data) == 0 {
		return ""
	}

	// cmdline is null-separated, replace nulls with spaces
	for i := range data {
		if data[i] == 0 {
			data[i] = ' '
		}
	}

	return strings.TrimSpace(string(data))
}

// lookupUsername returns username for a UID.
func (c *Collector) lookupUsername(uid int) string {
	if username, ok := c.uidCache[uid]; ok {
		return username
	}
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return strconv.Itoa(uid)
	}
	c.uidCache[uid] = u.Username
	return u.Username
}

//--------------

// walkPids calls fn for each process ID in /proc.
// It uses getdents64 directly for performance.
// Returns early if fn returns false.
func walkPids(fn func(int) bool) error {
	fd, err := syscall.Open("/proc", syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("open /proc: %w", err)
	}
	defer syscall.Close(fd)

	buf := make([]byte, 8192)
	for {
		n, err := syscall.Getdents(fd, buf)
		if err != nil {
			return fmt.Errorf("getdents: %w", err)
		}
		if n == 0 {
			break
		}

		if !parseDirents(buf[:n], fn) {
			break
		}
	}

	return nil
}

// parseDirents extracts PIDs from getdents64 buffer and calls fn.
// linux_dirent64 layout (all 64-bit Linux):
//
//	[0:8]   d_ino (u64)
//	[8:16]  d_off (s64)
//	[16:18] d_reclen (u16, little-endian)
//	[18]    d_type (u8)
//	[19:]   d_name (null-terminated string)
func parseDirents(buf []byte, fn func(int) bool) bool {
	pos := 0
	for pos+19 <= len(buf) {
		reclen := int(buf[pos+16]) | int(buf[pos+17])<<8
		if reclen < 19 || pos+reclen > len(buf) {
			break
		}

		dtype := buf[pos+18]
		// Only process directories (4=DT_DIR, 0=DT_UNKNOWN for compat)
		if dtype == 4 || dtype == 0 {
			nameStart := pos + 19
			nameEnd := pos + reclen

			// Find null terminator
			i := nameStart
			for i < nameEnd && buf[i] != 0 {
				i++
			}
			name := buf[nameStart:i]

			// Quick filter: PIDs start with digit 1-9
			if len(name) > 0 && name[0] >= '1' && name[0] <= '9' {
				if pid, ok := atoi(name); ok {
					if !fn(pid) {
						return false
					}
				}
			}
		}

		pos += reclen
	}
	return true
}

// atoi parses a decimal integer from bytes.
// It only handles non-negative values.
func atoi(b []byte) (int, bool) {
	if len(b) == 0 || len(b) > 10 {
		return 0, false
	}

	n := 0
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}

	return n, n > 0
}
