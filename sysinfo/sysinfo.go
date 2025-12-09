// Package sysinfo provides system resource monitoring using gopsutil.
package sysinfo

// Stats represents real-time system statistics.
type Stats struct {
	CPU     CPUStats   `json:"cpu"`
	Memory  MemStats   `json:"memory"`
	Network []NetStats `json:"network"`
	DiskIO  []DiskIO   `json:"disk_io"`
}

// CPUStats represents CPU usage statistics.
type CPUStats struct {
	Cores       []float64 `json:"cores"`       // Per-core usage percentage (0-100)
	Total       float64   `json:"total"`       // Aggregate usage percentage (0-100)
	Temperature float64   `json:"temperature"` // CPU temperature in Celsius (0 if unavailable)
	Frequency   float64   `json:"frequency"`   // Current frequency in MHz
	CoreCount   int       `json:"core_count"`  // Number of logical cores
}

// MemStats represents memory usage statistics.
type MemStats struct {
	Total     uint64  `json:"total"`      // Total RAM in bytes
	Used      uint64  `json:"used"`       // Used RAM in bytes
	Available uint64  `json:"available"`  // Available RAM in bytes
	Cached    uint64  `json:"cached"`     // Cached memory in bytes
	Buffers   uint64  `json:"buffers"`    // Buffer memory in bytes
	SwapTotal uint64  `json:"swap_total"` // Total swap in bytes
	SwapUsed  uint64  `json:"swap_used"`  // Used swap in bytes
	Percent   float64 `json:"percent"`    // Memory usage percentage (0-100)
}

// NetStats represents network interface statistics.
type NetStats struct {
	Name      string  `json:"name"`       // Interface name (e.g., "eth0")
	BytesIn   uint64  `json:"bytes_in"`   // Total bytes received
	BytesOut  uint64  `json:"bytes_out"`  // Total bytes transmitted
	SpeedIn   float64 `json:"speed_in"`   // Current receive rate (bytes/sec)
	SpeedOut  float64 `json:"speed_out"`  // Current transmit rate (bytes/sec)
	LinkSpeed uint64  `json:"link_speed"` // Link speed in Mbps (0 if unavailable)
	IsUp      bool    `json:"is_up"`      // Whether interface is up
}

// DiskIO represents disk I/O statistics.
type DiskIO struct {
	Device     string  `json:"device"`      // Device name (e.g., "sda")
	ReadBytes  uint64  `json:"read_bytes"`  // Total bytes read
	WriteBytes uint64  `json:"write_bytes"` // Total bytes written
	ReadSpeed  float64 `json:"read_speed"`  // Current read rate (bytes/sec)
	WriteSpeed float64 `json:"write_speed"` // Current write rate (bytes/sec)
}

// Process represents a running process.
type Process struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	Command    string  `json:"command"`
	User       string  `json:"user"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	MemRSS     uint64  `json:"mem_rss"`    // Resident set size in bytes
	State      string  `json:"state"`      // R=running, S=sleeping, etc.
	StartTime  int64   `json:"start_time"` // Unix timestamp
	Threads    int     `json:"threads"`    // Number of threads
}
