package zfs

// PoolStatus represents the health status of a pool.
type PoolStatus string

const (
	PoolOnline   PoolStatus = "ONLINE"
	PoolDegraded PoolStatus = "DEGRADED"
	PoolFaulted  PoolStatus = "FAULTED"
	PoolOffline  PoolStatus = "OFFLINE"
	PoolUnavail  PoolStatus = "UNAVAIL"
)

// Pool represents a ZFS storage pool.
type Pool struct {
	Name           string          `json:"name"`
	GUID           string          `json:"guid"`
	Size           uint64          `json:"size"`
	Allocated      uint64          `json:"allocated"`
	Free           uint64          `json:"free"`
	Frag           uint64          `json:"frag"` // Fragmentation percentage
	Health         PoolStatus      `json:"health"`
	VDevs          []VDevDetail    `json:"vdevs,omitempty"`
	DiskCount      int             `json:"disk_count"`
	Redundancy     int             `json:"redundancy"` // How many more disks can fail
	ScrubStatus    *ScrubStatus    `json:"scrub_status,omitempty"`
	ResilverStatus *ResilverStatus `json:"resilver_status,omitempty"`
}

// DatasetType represents the type of a dataset.
type DatasetType string

const (
	DatasetFilesystem DatasetType = "filesystem"
	DatasetVolume     DatasetType = "volume"
	DatasetSnapshot   DatasetType = "snapshot"
)

// Dataset represents a ZFS dataset.
type Dataset struct {
	Name          string      `json:"name"`
	Type          DatasetType `json:"type"`
	Pool          string      `json:"pool"` // Pool name extracted from dataset name
	Used          uint64      `json:"used"`
	Available     uint64      `json:"available"`
	Referenced    uint64      `json:"referenced"`
	Mountpoint    string      `json:"mountpoint"`
	Compression   string      `json:"compression"`
	Encryption    string      `json:"encryption"`
	Deduplication string      `json:"deduplication"`
	Quota         uint64      `json:"quota,omitempty"`
	Reservation   uint64      `json:"reservation,omitempty"`
}

// UseCaseTemplate represents predefined dataset configurations.
type UseCaseTemplate string

const (
	UseCaseGeneral      UseCaseTemplate = "general"
	UseCaseMedia        UseCaseTemplate = "media"
	UseCaseSurveillance UseCaseTemplate = "surveillance"
	UseCaseVM           UseCaseTemplate = "vm"
	UseCaseDatabase     UseCaseTemplate = "database"
)

// Snapshot represents a ZFS snapshot.
type Snapshot struct {
	Name       string `json:"name"`
	Dataset    string `json:"dataset"`
	CreatedAt  string `json:"created_at"`
	Used       uint64 `json:"used"`
	Referenced uint64 `json:"referenced"`
	Source     string `json:"source"` // "manual", "policy:daily", etc.
}

// ScrubAction represents scrub control actions.
type ScrubAction string

const (
	ScrubStart ScrubAction = "start"
	ScrubStop  ScrubAction = "stop"
	ScrubPause ScrubAction = "pause"
)

// ScrubStatus represents the status of a scrub operation.
type ScrubStatus struct {
	InProgress  bool    `json:"in_progress"`
	EndTime     *string `json:"end_time,omitempty"`
	Errors      int     `json:"errors"`
	DataScanned uint64  `json:"data_scanned"`
	DataToScan  uint64  `json:"data_to_scan"`
	ScanRate    uint64  `json:"scan_rate"` // bytes/sec
}

// VDevDetail represents detailed vdev information including disk status.
type VDevDetail struct {
	Name     string       `json:"name"`     // e.g., "mirror-0"
	Type     string       `json:"type"`     // mirror, raidz, raidz2, etc.
	Status   string       `json:"status"`   // ONLINE, DEGRADED, FAULTED
	Children []DiskDetail `json:"children"` // disks in this vdev
}

// DiskDetail represents a disk within a vdev.
type DiskDetail struct {
	Name      string `json:"name"`      // e.g., "sda"
	Path      string `json:"path"`      // e.g., "/dev/sda"
	Status    string `json:"status"`    // ONLINE, DEGRADED, FAULTED, OFFLINE
	Slot      string `json:"slot"`      // physical slot number if available
	Read      uint64 `json:"read"`      // read errors
	Write     uint64 `json:"write"`     // write errors
	Checksum  uint64 `json:"checksum"`  // checksum errors
	Replacing bool   `json:"replacing"` // is being replaced
}

// ResilverStatus represents the status of a resilver (rebuild) operation.
type ResilverStatus struct {
	InProgress   bool    `json:"in_progress"`
	PercentDone  float64 `json:"percent_done"`
	StartTime    int64   `json:"start_time"` // Unix timestamp, for frontend to calculate remaining time
	ScannedBytes uint64  `json:"scanned_bytes"`
	IssuedBytes  uint64  `json:"issued_bytes"` // bytes processed (for rate calculation)
	TotalBytes   uint64  `json:"total_bytes"`
	Rate         uint64  `json:"rate"` // bytes/sec
}

// PoolHealth represents pool health information for UI.
type PoolHealth struct {
	Status          PoolStatus `json:"status"`
	CanLoseMore     int        `json:"can_lose_more"`    // how many more disks can fail
	RiskLevel       string     `json:"risk_level"`       // "low", "medium", "high", "critical"
	RiskDescription string     `json:"risk_description"` // human-readable risk description
	Recommendation  string     `json:"recommendation"`   // what user should do
}

// CreatePoolRequest represents the request to create a new pool.
type CreatePoolRequest struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"` // List of disk paths (e.g., /dev/sda)
	Type    string   `json:"type"`    // mirror, raidz, raidz2, or empty for stripe
}

// CreateSnapshotRequest represents a request to create a snapshot.
type CreateSnapshotRequest struct {
	Dataset string `json:"dataset"` // pool/dataset name
	Name    string `json:"name"`    // snapshot name (without @)
}
