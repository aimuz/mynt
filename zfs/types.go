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
	Name      string     `json:"name"`
	GUID      string     `json:"guid"`
	Size      uint64     `json:"size"`
	Allocated uint64     `json:"allocated"`
	Free      uint64     `json:"free"`
	Frag      uint64     `json:"frag"` // Fragmentation percentage
	Health    PoolStatus `json:"health"`
	AltRoot   string     `json:"altroot"`
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
}

// CreatePoolRequest represents the request to create a new pool.
type CreatePoolRequest struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"` // List of disk paths (e.g., /dev/sda)
	Type    string   `json:"type"`    // mirror, raidz, raidz2, or empty for stripe
}
