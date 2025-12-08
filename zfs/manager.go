package zfs

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

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
	zpools, err := gozfs.ListZpools()
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	pools := make([]Pool, 0, len(zpools))
	for _, zp := range zpools {
		pools = append(pools, fromGozfsPool(zp))
	}

	return pools, nil
}

// ListDatasets lists all datasets.
func (m *Manager) ListDatasets(ctx context.Context) ([]Dataset, error) {
	fsDatasets, err := gozfs.Filesystems("")
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	VolDatasets, err := gozfs.Volumes("")
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	gozfsDatasets := slices.Concat(fsDatasets, VolDatasets)
	slices.SortFunc(gozfsDatasets, func(a, b *gozfs.Dataset) int {
		return strings.Compare(a.Name, b.Name)
	})

	datasets := make([]Dataset, 0, len(gozfsDatasets))
	for _, gd := range gozfsDatasets {
		ds := fromGozfsDataset(gd)

		// Get encryption and dedup properties
		if enc, err := gd.GetProperty("encryption"); err == nil {
			ds.Encryption = enc
		}
		if dedup, err := gd.GetProperty("dedup"); err == nil {
			ds.Deduplication = dedup
		}

		datasets = append(datasets, ds)
	}

	return datasets, nil
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

// GetPool gets comprehensive details of a single pool from a single zpool status call.
func (m *Manager) GetPool(ctx context.Context, name string) (*Pool, error) {
	// Execute zpool status -p -j for all pool info
	out, err := m.exec.Output(ctx, "zpool", "status", "-p", "-j", name)
	if err != nil {
		return nil, fmt.Errorf("get pool status %s: %w", name, err)
	}

	var status ZpoolStatusJSON
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("parse pool status JSON: %w", err)
	}

	poolJSON, ok := status.Pools[name]
	if !ok {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	// Parse vdevs from JSON
	vdevs := parseVDevsFromJSON(poolJSON.VDevs)

	// Build pool from JSON data
	pool := poolFromJSON(name, poolJSON, vdevs)

	// Parse scan status (scrub/resilver)
	pool.ScrubStatus = parseScrubFromJSON(poolJSON.ScanStats)
	pool.ResilverStatus = parseResilverFromJSON(poolJSON.ScanStats)

	return &pool, nil
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
		for _, v := range root.VDevs {
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
	for _, child := range children {
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
