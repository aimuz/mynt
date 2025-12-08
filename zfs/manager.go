package zfs

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	gozfs "github.com/mistifyio/go-zfs/v4"
	"go.aimuz.me/mynt/logger"
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

// ScrubStatus gets the scrub status of a pool.
func (m *Manager) ScrubStatus(ctx context.Context, poolName string) (string, error) {
	out, err := m.exec.Output(ctx, "zpool", "status", poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get scrub status: %w", err)
	}
	return string(out), nil
}

// GetPool gets details of a single pool.
func (m *Manager) GetPool(ctx context.Context, name string) (*Pool, error) {
	zpool, err := gozfs.GetZpool(name)
	if err != nil {
		return nil, fmt.Errorf("get pool %s: %w", name, err)
	}
	pool := fromGozfsPool(zpool)

	// Parse vdev structure from zpool status
	vdevs, err := m.GetPoolVDevs(ctx, name)
	if err != nil {
		// Log but don't fail - vdev info is supplementary
		logger.Warn("failed to parse vdev structure", "pool", name, "error", err)
	} else {
		pool.VDevs = make([]VDev, len(vdevs))
		diskCount := 0
		for i, vd := range vdevs {
			pool.VDevs[i] = VDev{
				Type:   vd.Type,
				Disks:  make([]string, len(vd.Children)),
				Status: vd.Status,
			}
			for j, child := range vd.Children {
				pool.VDevs[i].Disks[j] = child.Path
			}
			diskCount += len(vd.Children)
		}
		pool.DiskCount = diskCount
		pool.Redundancy = calculateRedundancy(vdevs)
	}

	return &pool, nil
}

// GetPoolVDevs parses the zpool status output to get vdev structure.
func (m *Manager) GetPoolVDevs(ctx context.Context, poolName string) ([]VDevDetail, error) {
	out, err := m.exec.Output(ctx, "zpool", "status", poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool status: %w", err)
	}

	return parsePoolStatus(string(out), poolName)
}

// ReplaceDisk replaces a disk in a pool.
func (m *Manager) ReplaceDisk(ctx context.Context, poolName, oldDisk, newDisk string) error {
	_, err := m.exec.Output(ctx, "zpool", "replace", "-f", poolName, oldDisk, newDisk)
	if err != nil {
		return fmt.Errorf("replace disk %s with %s in pool %s: %w", oldDisk, newDisk, poolName, err)
	}
	return nil
}

// GetResilverStatus gets the resilver (rebuild) status of a pool.
func (m *Manager) GetResilverStatus(ctx context.Context, poolName string) (*ResilverStatus, error) {
	out, err := m.exec.Output(ctx, "zpool", "status", poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool status: %w", err)
	}
	return parseResilverStatus(string(out)), nil
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

// parsePoolStatus parses zpool status output to extract vdev structure.
func parsePoolStatus(output, poolName string) ([]VDevDetail, error) {
	lines := strings.Split(output, "\n")
	var vdevs []VDevDetail
	var currentVDev *VDevDetail
	inConfig := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for the config section
		if strings.HasPrefix(trimmed, "config:") {
			inConfig = true
			continue
		}

		if !inConfig {
			continue
		}

		// Skip empty lines and header
		if trimmed == "" || strings.HasPrefix(trimmed, "NAME") {
			continue
		}

		// Stop at errors section
		if strings.HasPrefix(trimmed, "errors:") {
			break
		}

		// Parse the line
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		status := fields[1]

		// Skip the pool name line
		if name == poolName {
			continue
		}

		// Check if this is a vdev type (mirror, raidz, etc.)
		if strings.HasPrefix(name, "mirror") ||
			strings.HasPrefix(name, "raidz") ||
			strings.HasPrefix(name, "spare") ||
			strings.HasPrefix(name, "log") ||
			strings.HasPrefix(name, "cache") {

			// Determine vdev type
			vdevType := "stripe"
			if strings.HasPrefix(name, "mirror") {
				vdevType = "mirror"
			} else if strings.HasPrefix(name, "raidz3") {
				vdevType = "raidz3"
			} else if strings.HasPrefix(name, "raidz2") {
				vdevType = "raidz2"
			} else if strings.HasPrefix(name, "raidz") {
				vdevType = "raidz"
			}

			currentVDev = &VDevDetail{
				Name:     name,
				Type:     vdevType,
				Status:   status,
				Children: []DiskDetail{},
			}
			vdevs = append(vdevs, *currentVDev)
			continue
		}

		// This is a disk
		// Note: name could be just device name (e.g. "sda") or an absolute path (e.g. "/dev/disk/by-id/...")
		path := name
		if !strings.HasPrefix(name, "/") {
			path = filepath.Join("/dev", name)
		}
		disk := DiskDetail{
			Name:      name,
			Path:      path,
			Status:    status,
			Replacing: strings.Contains(line, "replacing"),
		}

		// Parse error counts if present (fields: NAME STATE READ WRITE CKSUM)
		if len(fields) >= 5 {
			disk.Read, _ = strconv.ParseUint(fields[2], 10, 64)
			disk.Write, _ = strconv.ParseUint(fields[3], 10, 64)
			disk.Checksum, _ = strconv.ParseUint(fields[4], 10, 64)
		}

		if currentVDev != nil && len(vdevs) > 0 {
			vdevs[len(vdevs)-1].Children = append(vdevs[len(vdevs)-1].Children, disk)
		} else {
			// Single disk pool (stripe)
			vdevs = append(vdevs, VDevDetail{
				Name:     "stripe",
				Type:     "stripe",
				Status:   status,
				Children: []DiskDetail{disk},
			})
		}
	}

	return vdevs, nil
}

// parseResilverStatus parses zpool status output to extract resilver progress.
// TODO(zfs2.3): ZFS 2.3+ supports `zpool status -j` JSON output, which would
// simplify parsing significantly. Consider switching when ZFS 2.3 is widely adopted.
func parseResilverStatus(output string) *ResilverStatus {
	status := &ResilverStatus{
		InProgress: false,
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		// Look for resilver in progress
		if strings.Contains(line, "resilver in progress") {
			status.InProgress = true
			continue
		}

		// Parse scan line: "scanned 112G of 1.81T at 1.14G/s, 0h24m to go"
		// or "scanned 50.0G resilvered 25.0G at 100M/s, 0:30:00 to go"
		if strings.Contains(line, "scanned") && strings.Contains(line, "at") {
			status.ScannedBytes, status.TotalBytes, status.Rate = parseScanLine(line)
		}

		// Parse progress line: "... 50.0% done, ..."
		if strings.Contains(line, "% done") {
			for _, field := range strings.Fields(line) {
				if strings.HasSuffix(field, "%") {
					pctStr := strings.TrimSuffix(field, "%")
					if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
						status.PercentDone = pct
					}
					break
				}
			}
		}

		// Parse time remaining: "... 1h30m to go" or "... 0:30:00 to go"
		if idx := strings.Index(line, " to go"); idx > 0 {
			// Find the time part before "to go"
			parts := strings.Fields(line[:idx])
			if len(parts) > 0 {
				status.TimeRemaining = parts[len(parts)-1]
			}
		}
	}

	return status
}

// parseScanLine parses a scan line like "112G scanned of 1.81T at 1.14G/s"
// Returns scanned bytes, total bytes, and rate in bytes/sec.
func parseScanLine(line string) (scanned, total, rate uint64) {
	fields := strings.Fields(line)
	for i, f := range fields {
		// Look for "XXX scanned" - size comes BEFORE keyword
		if f == "scanned" && i > 0 {
			scanned = parseSize(fields[i-1])
		}
		// Look for "of XXX"
		if f == "of" && i+1 < len(fields) {
			total = parseSize(fields[i+1])
		}
		// Look for "at XXX/s" or "at XXX/s,"
		if f == "at" && i+1 < len(fields) {
			rateStr := fields[i+1]
			rateStr = strings.TrimSuffix(rateStr, ",")
			rateStr = strings.TrimSuffix(rateStr, "/s")
			rate = parseSize(rateStr)
		}
	}
	return
}

// parseSize parses size strings like "112G", "1.81T", "100M" to bytes.
func parseSize(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := uint64(1)
	lastChar := s[len(s)-1]

	switch lastChar {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'T', 't':
		multiplier = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'P', 'p':
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return uint64(val * float64(multiplier))
}
