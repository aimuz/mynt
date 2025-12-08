package zfs

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strconv"

	gozfs "github.com/mistifyio/go-zfs/v4"
	"go.aimuz.me/mynt/sysexec"
)

// Manager handles ZFS operations.
type Manager struct {
	exec sysexec.Executor
}

// NewManager creates a new ZFS manager.
func NewManager() *Manager {
	return &Manager{exec: sysexec.NewExecutor()}
}

// ListPools lists all imported ZFS pools.
func (m *Manager) ListPools(ctx context.Context) ([]Pool, error) {
	return m.listPools(ctx)
}

// GetPool gets comprehensive details of a single pool.
func (m *Manager) GetPool(ctx context.Context, name string) (*Pool, error) {
	pools, err := m.listPools(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(pools) == 0 {
		return nil, fmt.Errorf("pool %s not found", name)
	}
	return &pools[0], nil
}

// listPools is the internal implementation for listing pools.
// If names are provided, only those pools are queried.
func (m *Manager) listPools(ctx context.Context, names ...string) ([]Pool, error) {
	args := []string{"status", "-p", "-j"}
	args = append(args, names...)

	out, err := m.exec.Output(ctx, "zpool", args...)
	if err != nil {
		return nil, fmt.Errorf("zpool status: %w", err)
	}

	var status ZpoolStatusJSON
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("parse zpool status: %w", err)
	}

	pools := make([]Pool, 0, len(status.Pools))
	for name, pj := range status.Pools {
		pools = append(pools, buildPool(name, pj))
	}
	return pools, nil
}

// buildPool constructs a Pool from JSON data.
func buildPool(name string, pj *PoolJSON) Pool {
	vdevs := parseVDevsFromJSON(pj.VDevs)
	pool := poolFromJSON(name, pj, vdevs)
	pool.ScrubStatus = parseScrubFromJSON(pj.ScanStats)
	pool.ResilverStatus = parseResilverFromJSON(pj.ScanStats)
	return pool
}

// listDatasets is the internal implementation for listing datasets.
// If names are provided, only those datasets are queried.
func (m *Manager) listDatasets(ctx context.Context, names ...string) ([]Dataset, error) {
	args := []string{"list", "-j", "-p", "-t", "filesystem,volume",
		"-o", "name,type,used,available,referenced,mountpoint,compression,encryption,dedup,quota,reservation,volsize,usedbydataset"}
	args = append(args, names...)

	out, err := m.exec.Output(ctx, "zfs", args...)
	if err != nil {
		return nil, fmt.Errorf("zfs list: %w", err)
	}

	var listJSON ZFSListJSON
	if err := json.Unmarshal(out, &listJSON); err != nil {
		return nil, fmt.Errorf("parse zfs list: %w", err)
	}

	datasets := make([]Dataset, 0, len(listJSON.Datasets))
	for _, dj := range sortMapIter(listJSON.Datasets) {
		datasets = append(datasets, buildDataset(dj))
	}

	return datasets, nil
}

// buildDataset constructs a Dataset from JSON data.
func buildDataset(dj *DatasetListJSON) Dataset {
	dsType := DatasetFilesystem
	if dj.Type == "VOLUME" {
		dsType = DatasetVolume
	}

	getPropValue := func(key string) string {
		if dj.Properties == nil {
			return ""
		}
		if p, ok := dj.Properties[key]; ok && p != nil {
			return p.Value
		}
		return ""
	}

	used := parseUint(getPropValue("used"))
	if dsType == DatasetVolume {
		used = parseUint(getPropValue("usedbydataset"))
	}

	quota := parseUint(getPropValue("quota"))
	if dsType == DatasetVolume {
		quota = parseUint(getPropValue("volsize"))
	}

	return Dataset{
		Name:          dj.Name,
		Pool:          dj.Pool,
		Type:          dsType,
		Used:          used,
		Available:     parseUint(getPropValue("available")),
		Referenced:    parseUint(getPropValue("referenced")),
		Mountpoint:    getPropValue("mountpoint"),
		Compression:   getPropValue("compression"),
		Encryption:    getPropValue("encryption"),
		Deduplication: getPropValue("dedup"),
		Quota:         quota,
		Reservation:   parseUint(getPropValue("reservation")),
	}
}

// CreatePool creates a new ZFS pool.
func (m *Manager) CreatePool(ctx context.Context, req CreatePoolRequest) error {
	// Add optional properties (none for now, but structure is ready)
	properties := map[string]string{
		"mountpoint": fmt.Sprintf("/mnt/%s", req.Name),
	}

	// Build vdev args
	vdevArgs := make([]string, 0)
	if req.Type != "" {
		vdevArgs = append(vdevArgs, req.Type)
	}
	vdevArgs = append(vdevArgs, req.Devices...)

	_, err := gozfs.CreateZpool(req.Name, properties, vdevArgs...)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	return nil
}

// DestroyPool destroys a ZFS pool.
func (m *Manager) DestroyPool(ctx context.Context, name string) error {
	zpool, err := gozfs.GetZpool(name)
	if err != nil {
		return fmt.Errorf("failed to get pool: %w", err)
	}

	if err := zpool.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy pool: %w", err)
	}

	return nil
}

// Scrub starts a scrub operation on a pool.
// Note: go-zfs/v4 doesn't provide scrub functionality, so we implement it ourselves.
func (m *Manager) Scrub(ctx context.Context, poolName string) error {
	_, err := m.exec.Output(ctx, "zpool", "scrub", poolName)
	if err != nil {
		return fmt.Errorf("failed to start scrub: %w", err)
	}
	return nil
}

// poolFromJSON builds a Pool from JSON data.
func poolFromJSON(name string, p *PoolJSON, vdevs []VDevDetail) Pool {
	// Find root vdev to get size/allocated
	var size, allocated uint64
	for _, v := range p.VDevs {
		if v.VDevType == "root" {
			size = parseUint(v.TotalSpace)
			allocated = parseUint(v.AllocSpace)
			break
		}
	}

	// Count disks and calculate redundancy
	diskCount := 0
	for _, vd := range vdevs {
		diskCount += len(vd.Children)
	}

	return Pool{
		Name:       name,
		GUID:       p.PoolGUID,
		Size:       size,
		Allocated:  allocated,
		Free:       size - allocated,
		Health:     PoolStatus(p.State),
		VDevs:      vdevs,
		DiskCount:  diskCount,
		Redundancy: calculateRedundancy(vdevs),
	}
}

// parseScrubFromJSON converts ScanStatsJSON to ScrubStatus.
func parseScrubFromJSON(scan *ScanStatsJSON) *ScrubStatus {
	if scan == nil || scan.Function != "SCRUB" {
		return nil
	}

	status := &ScrubStatus{
		InProgress:  scan.State == "SCANNING",
		Errors:      int(parseUint(scan.Errors)),
		DataScanned: parseUint(scan.Examined),
		DataToScan:  parseUint(scan.ToExamine),
		ScanRate:    parseUint(scan.BytesPerScan),
	}

	if scan.State == "FINISHED" && scan.EndTime != "" {
		status.EndTime = &scan.EndTime
	}

	return status
}

// ReplaceDisk replaces a disk in a pool.
func (m *Manager) ReplaceDisk(ctx context.Context, poolName, oldDisk, newDisk string) error {
	_, err := m.exec.Output(ctx, "zpool", "replace", "-f", poolName, oldDisk, newDisk)
	if err != nil {
		return fmt.Errorf("replace disk %s with %s in pool %s: %w", oldDisk, newDisk, poolName, err)
	}
	return nil
}

// calculateRedundancy determines how many more disks can fail.
// Returns the minimum number of additional disks that can fail before data loss.
// Returns 0 if already degraded with no more redundancy.
func calculateRedundancy(vdevs []VDevDetail) int {
	if len(vdevs) == 0 {
		return 0
	}

	minRedundancy := -1
	for _, vdev := range vdevs {
		r := 0
		switch vdev.Type {
		case "mirror":
			// Mirror can lose all but one disk
			online := 0
			for _, d := range vdev.Children {
				if d.Status == "ONLINE" {
					online++
				}
			}
			r = online - 1
		case "raidz":
			r = 1
		case "raidz2":
			r = 2
		case "raidz3":
			r = 3
		default:
			// Single disk or stripe
			r = 0
		}

		// Account for already failed disks
		for _, d := range vdev.Children {
			if d.Status != "ONLINE" {
				r--
			}
		}
		if r < 0 {
			r = 0
		}

		if minRedundancy < 0 || r < minRedundancy {
			minRedundancy = r
		}
	}

	if minRedundancy < 0 {
		return 0
	}
	return minRedundancy
}

// parseVDevsFromJSON converts JSON vdevs to VDevDetail slice.
// It iterates through the tree structure: root -> vdev (mirror/raidz/disk) -> disk
func parseVDevsFromJSON(jsonVDevs map[string]*Vdev) []VDevDetail {
	var vdevs []VDevDetail
	for _, root := range jsonVDevs {
		if root.VDevType != "root" {
			continue
		}
		for _, v := range sortMapIter(root.VDevs) {
			vdevs = append(vdevs, vdevDetailFromVdev(v))
		}
	}
	return vdevs
}

// vdevDetailFromVdev converts a single Vdev node to VDevDetail.
func vdevDetailFromVdev(v *Vdev) VDevDetail {
	vdev := VDevDetail{
		Name:   v.Name,
		Type:   vdevTypeFromJSON(v.VDevType),
		Status: v.State,
	}
	if len(v.VDevs) == 0 {
		// Single disk (stripe pool)
		vdev.Children = []DiskDetail{diskDetailFromVdev(v, false)}
		return vdev
	}
	// Mirror/raidz with child disks
	vdev.Children = collectChildDisks(v.VDevs)
	return vdev
}

// collectChildDisks extracts DiskDetail from child vdevs, handling "replacing" vdevs.
func collectChildDisks(children map[string]*Vdev) []DiskDetail {
	var disks []DiskDetail
	for _, child := range sortMapIter(children) {
		if child.VDevType == "replacing" {
			// Replacing vdev contains old and new disk as children
			for _, d := range child.VDevs {
				disks = append(disks, diskDetailFromVdev(d, true))
			}
		} else {
			disks = append(disks, diskDetailFromVdev(child, false))
		}
	}
	return disks
}

// diskDetailFromVdev creates a DiskDetail from a Vdev node.
func diskDetailFromVdev(v *Vdev, replacing bool) DiskDetail {
	return DiskDetail{
		Name:      v.Name,
		Path:      v.Path,
		Status:    v.State,
		Read:      parseUint(v.ReadErrors),
		Write:     parseUint(v.WriteErrors),
		Checksum:  parseUint(v.ChecksumErrors),
		Replacing: replacing,
	}
}

// vdevTypeFromJSON normalizes vdev_type from JSON to our internal type.
func vdevTypeFromJSON(jsonType string) string {
	switch jsonType {
	case "mirror":
		return "mirror"
	case "raidz1", "raidz":
		return "raidz"
	case "raidz2":
		return "raidz2"
	case "raidz3":
		return "raidz3"
	case "disk":
		return "stripe"
	default:
		return jsonType
	}
}

// parseResilverFromJSON converts ScanStatsJSON to ResilverStatus.
func parseResilverFromJSON(scan *ScanStatsJSON) *ResilverStatus {
	status := &ResilverStatus{
		InProgress: false,
	}

	if scan == nil {
		return status
	}

	// Check if this is an active resilver
	if scan.Function == "RESILVER" && scan.State == "SCANNING" {
		status.InProgress = true
		status.StartTime = int64(parseUint(scan.PassStart))
		status.ScannedBytes = parseUint(scan.Examined)
		status.IssuedBytes = parseUint(scan.Issued)
		status.TotalBytes = parseUint(scan.ToExamine)
		status.Rate = parseUint(scan.BytesPerScan)

		// Calculate percent done
		if status.TotalBytes > 0 {
			status.PercentDone = float64(status.ScannedBytes) / float64(status.TotalBytes) * 100
		}
	}

	return status
}

// parseUint safely parses a string to uint64, returning 0 on error.
func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// sortMapIter returns an iterator that yields map entries in sorted key order.
// The iterator conforms to iter.Seq2 and can be used with range loops.
func sortMapIter[K string, T any](m map[K]T) iter.Seq2[K, T] {
	// Collect and sort keys first to ensure deterministic iteration order.
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	return func(yield func(K, T) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}
