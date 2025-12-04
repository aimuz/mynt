package zfs

import (
	"context"
	"fmt"
	"slices"
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

// ScrubStatus gets the scrub status of a pool.
func (m *Manager) ScrubStatus(ctx context.Context, poolName string) (string, error) {
	out, err := m.exec.Output(ctx, "zpool", "status", poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get scrub status: %w", err)
	}
	return string(out), nil
}
