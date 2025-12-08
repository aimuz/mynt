package zfs

// ZpoolStatusJSON represents the JSON output of `zpool status -j`.
type ZpoolStatusJSON struct {
	OutputVersion OutputVersion        `json:"output_version"`
	Pools         map[string]*PoolJSON `json:"pools"`
}

// OutputVersion contains version info for the JSON output format.
type OutputVersion struct {
	Command   string `json:"command"`
	VersMajor int    `json:"vers_major"`
	VersMinor int    `json:"vers_minor"`
}

// PoolJSON represents a pool in the JSON output.
type PoolJSON struct {
	Name       string               `json:"name"`
	State      string               `json:"state"`
	PoolGUID   string               `json:"pool_guid"`
	TXG        string               `json:"txg"`
	SPAVersion string               `json:"spa_version"`
	ZPLVersion string               `json:"zpl_version"`
	ScanStats  *ScanStatsJSON       `json:"scan_stats,omitempty"`
	VDevs      map[string]*VDevJSON `json:"vdevs"`
	ErrorCount string               `json:"error_count"`
}

// ScanStatsJSON represents scan (scrub/resilver) statistics.
type ScanStatsJSON struct {
	Function         string `json:"function"`           // "SCRUB" or "RESILVER"
	State            string `json:"state"`              // "SCANNING", "FINISHED", "CANCELED"
	StartTime        string `json:"start_time"`         // Human readable start time
	EndTime          string `json:"end_time,omitempty"` // Human readable end time (if finished)
	ToExamine        string `json:"to_examine"`         // Total bytes to examine (e.g., "1.81T")
	Examined         string `json:"examined"`           // Bytes examined so far (e.g., "112G")
	Skipped          string `json:"skipped"`            // Bytes skipped
	Processed        string `json:"processed"`          // Bytes processed
	Errors           string `json:"errors"`             // Number of errors
	BytesPerScan     string `json:"bytes_per_scan"`     // Scan rate
	PassStart        string `json:"pass_start"`         // Unix timestamp of pass start
	ScrubPause       string `json:"scrub_pause"`        // Pause status
	ScrubSpentPaused string `json:"scrub_spent_paused"` // Time spent paused
	Issued           string `json:"issued"`             // Bytes issued
}

// VDevJSON represents a vdev in the JSON output.
// VDevs are nested recursively (root -> mirror/raidz -> disk).
type VDevJSON struct {
	Name           string               `json:"name"`
	VDevType       string               `json:"vdev_type"` // "root", "mirror", "raidz", "raidz2", "raidz3", "disk"
	GUID           string               `json:"guid"`
	Path           string               `json:"path,omitempty"` // Device path (for disk vdevs)
	Class          string               `json:"class"`          // "normal", "log", "cache", "spare"
	State          string               `json:"state"`          // "ONLINE", "DEGRADED", "FAULTED", "OFFLINE"
	AllocSpace     string               `json:"alloc_space"`
	TotalSpace     string               `json:"total_space"`
	DefSpace       string               `json:"def_space"`
	PhysSpace      string               `json:"phys_space"`
	RepDevSize     string               `json:"rep_dev_size,omitempty"`
	ReadErrors     string               `json:"read_errors"`
	WriteErrors    string               `json:"write_errors"`
	ChecksumErrors string               `json:"checksum_errors"`
	SlowIOs        string               `json:"slow_ios,omitempty"`
	VDevs          map[string]*VDevJSON `json:"vdevs,omitempty"` // Nested child vdevs
}
