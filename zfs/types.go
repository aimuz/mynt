package zfs

import (
	"cmp"

	gozfs "github.com/mistifyio/go-zfs/v4"
)

// PoolStatus represents the health status of a pool.
type PoolStatus string

const (
	PoolOnline   PoolStatus = "ONLINE"
	PoolDegraded PoolStatus = "DEGRADED"
	PoolFaulted  PoolStatus = "FAULTED"
	PoolOffline  PoolStatus = "OFFLINE"
	PoolUnavail  PoolStatus = "UNAVAIL"
)

// VDev represents a virtual device in a pool.
type VDev struct {
	Type   string   `json:"type"`   // "mirror", "raidz", "raidz2", "raidz3", etc.
	Disks  []string `json:"disks"`  // device paths
	Status string   `json:"status"` // ONLINE, DEGRADED, FAULTED, etc.
}

// Pool represents a ZFS storage pool.
type Pool struct {
	Name            string     `json:"name"`
	GUID            string     `json:"guid"`
	Size            uint64     `json:"size"`
	Allocated       uint64     `json:"allocated"`
	Free            uint64     `json:"free"`
	Frag            uint64     `json:"frag"` // Fragmentation percentage
	Health          PoolStatus `json:"health"`
	AltRoot         string     `json:"altroot"`
	VDevs           []VDev     `json:"vdevs,omitempty"`
	DiskCount       int        `json:"disk_count"`
	RedundancyLevel string     `json:"redundancy_level"` // "可坏 1 盘", "可坏 2 盘", etc.
	LastScrub       *string    `json:"last_scrub,omitempty"`
	ScrubInProgress bool       `json:"scrub_in_progress"`
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

// fromGozfsPool converts a go-zfs Zpool to our Pool type.
func fromGozfsPool(z *gozfs.Zpool) Pool {
	return Pool{
		Name:      z.Name,
		GUID:      "", // go-zfs doesn't provide GUID in the same way
		Size:      z.Size,
		Allocated: z.Allocated,
		Free:      z.Free,
		Frag:      z.Fragmentation,
		Health:    PoolStatus(z.Health),
		AltRoot:   "", // go-zfs doesn't provide AltRoot
	}
}

// fromGozfsDataset converts a go-zfs Dataset to our Dataset type.
func fromGozfsDataset(d *gozfs.Dataset) Dataset {
	used := d.Used
	if d.Type == gozfs.DatasetVolume {
		used = d.Usedbydataset
	}
	return Dataset{
		Name:          d.Name,
		Type:          DatasetType(d.Type),
		Used:          used,
		Available:     d.Avail,
		Referenced:    d.Referenced,
		Mountpoint:    d.Mountpoint,
		Compression:   d.Compression,
		Encryption:    "", // Need to get from properties
		Deduplication: "", // Need to get from properties
		Quota:         cmp.Or(d.Quota, d.Volsize),
	}
}
