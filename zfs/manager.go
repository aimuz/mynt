package zfs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

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
// It runs `zpool list -H -p -o name,guid,size,alloc,free,frag,health,altroot`
func (m *Manager) ListPools(ctx context.Context) ([]Pool, error) {
	args := []string{"list", "-H", "-p", "-o", "name,guid,size,alloc,free,frag,health,altroot"}
	out, err := m.exec.Output(ctx, "zpool", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	var pools []Pool
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		size, _ := strconv.ParseUint(fields[2], 10, 64)
		alloc, _ := strconv.ParseUint(fields[3], 10, 64)
		free, _ := strconv.ParseUint(fields[4], 10, 64)
		frag, _ := strconv.ParseUint(strings.TrimSuffix(fields[5], "%"), 10, 64)

		pools = append(pools, Pool{
			Name:      fields[0],
			GUID:      fields[1],
			Size:      size,
			Allocated: alloc,
			Free:      free,
			Frag:      frag,
			Health:    PoolStatus(fields[6]),
			AltRoot:   fields[7],
		})
	}

	return pools, nil
}

// ListDatasets lists all datasets.
// It runs `zfs list -H -p -o name,type,used,avail,refer,mountpoint,compression,encryption,dedup`
func (m *Manager) ListDatasets(ctx context.Context) ([]Dataset, error) {
	args := []string{"list", "-H", "-p", "-o", "name,type,used,avail,refer,mountpoint,compression,encryption,dedup"}
	out, err := m.exec.Output(ctx, "zfs", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	var datasets []Dataset
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) < 9 {
			continue
		}

		used, _ := strconv.ParseUint(fields[2], 10, 64)
		avail, _ := strconv.ParseUint(fields[3], 10, 64)
		refer, _ := strconv.ParseUint(fields[4], 10, 64)

		datasets = append(datasets, Dataset{
			Name:          fields[0],
			Type:          DatasetType(fields[1]),
			Used:          used,
			Available:     avail,
			Referenced:    refer,
			Mountpoint:    fields[5],
			Compression:   fields[6],
			Encryption:    fields[7],
			Deduplication: fields[8],
		})
	}

	return datasets, nil
}

// CreatePool creates a new ZFS pool.
// It runs `zpool create <name> <type> <devices...>`
func (m *Manager) CreatePool(ctx context.Context, req CreatePoolRequest) error {
	args := []string{"create", "-f", req.Name} // -f to force if needed (be careful in prod)
	if req.Type != "" {
		args = append(args, req.Type)
	}
	args = append(args, req.Devices...)

	_, err := m.exec.Output(ctx, "zpool", args...)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}
	return nil
}
