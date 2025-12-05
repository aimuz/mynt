package zfs

import (
	"context"
	"fmt"
	"strings"
	"time"

	gozfs "github.com/mistifyio/go-zfs/v4"
)

// CreateSnapshot creates a new ZFS snapshot.
func (m *Manager) CreateSnapshot(ctx context.Context, req CreateSnapshotRequest) (*Snapshot, error) {
	if req.Dataset == "" {
		return nil, fmt.Errorf("dataset name is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("snapshot name is required")
	}

	// Ensure snapshot name doesn't contain '@'
	snapshotName := strings.TrimPrefix(req.Name, "@")
	fullName := fmt.Sprintf("%s@%s", req.Dataset, snapshotName)

	dataset, err := gozfs.GetDataset(req.Dataset)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %s: %w", req.Dataset, err)
	}

	gozfsSnapshot, err := dataset.Snapshot(snapshotName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	snapshot := &Snapshot{
		Name:       fullName,
		Dataset:    req.Dataset,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Used:       gozfsSnapshot.Used,
		Referenced: gozfsSnapshot.Referenced,
		Source:     "manual",
	}

	return snapshot, nil
}

// ListSnapshots returns all snapshots for a specific dataset.
func (m *Manager) ListSnapshots(ctx context.Context, datasetName string) ([]Snapshot, error) {
	if datasetName == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	dataset, err := gozfs.GetDataset(datasetName)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %s: %w", datasetName, err)
	}

	gozfsSnapshots, err := dataset.Snapshots()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	snapshots := make([]Snapshot, 0, len(gozfsSnapshots))
	for _, snap := range gozfsSnapshots {
		// Parse creation time from properties if available
		createdAt := time.Now().Format(time.RFC3339)
		if prop, err := snap.GetProperty("creation"); err == nil {
			if t, err := parseZFSTimestamp(prop); err == nil {
				createdAt = t.Format(time.RFC3339)
			}
		}

		// Detect source from snapshot name
		// Auto snapshots use format: auto-{policyName}-{timestamp}
		source := "manual"
		parts := strings.Split(snap.Name, "@")
		if len(parts) == 2 {
			snapName := parts[1]
			if strings.HasPrefix(snapName, "auto-") {
				// Extract policy name: auto-{policyName}-{timestamp}
				// Remove "auto-" prefix, then find the last "-" before timestamp
				rest := strings.TrimPrefix(snapName, "auto-")
				// Timestamp format is YYYYMMDD-HHMMSS (15 chars)
				if len(rest) > 16 {
					policyName := rest[:len(rest)-16] // Remove "-YYYYMMDD-HHMMSS"
					source = "policy:" + policyName
				} else {
					source = "policy:auto"
				}
			}
		}

		snapshots = append(snapshots, Snapshot{
			Name:       snap.Name,
			Dataset:    datasetName,
			CreatedAt:  createdAt,
			Used:       snap.Used,
			Referenced: snap.Referenced,
			Source:     source,
		})
	}

	return snapshots, nil
}

// DestroySnapshot destroys a ZFS snapshot.
func (m *Manager) DestroySnapshot(ctx context.Context, snapshotName string) error {
	if snapshotName == "" {
		return fmt.Errorf("snapshot name is required")
	}

	if !strings.Contains(snapshotName, "@") {
		return fmt.Errorf("invalid snapshot name format (expected dataset@snapshot)")
	}

	snapshot, err := gozfs.GetDataset(snapshotName)
	if err != nil {
		return fmt.Errorf("snapshot not found: %s: %w", snapshotName, err)
	}

	if err := snapshot.Destroy(gozfs.DestroyDefault); err != nil {
		return fmt.Errorf("failed to destroy snapshot: %w", err)
	}

	return nil
}

// RollbackSnapshot rolls back a dataset to a specific snapshot.
func (m *Manager) RollbackSnapshot(ctx context.Context, snapshotName string) error {
	if snapshotName == "" {
		return fmt.Errorf("snapshot name is required")
	}

	if !strings.Contains(snapshotName, "@") {
		return fmt.Errorf("invalid snapshot name format (expected dataset@snapshot)")
	}

	snapshot, err := gozfs.GetDataset(snapshotName)
	if err != nil {
		return fmt.Errorf("snapshot not found: %s: %w", snapshotName, err)
	}

	if err := snapshot.Rollback(false); err != nil {
		return fmt.Errorf("failed to rollback snapshot: %w", err)
	}

	return nil
}

// CloneSnapshot creates a clone from a snapshot.
func (m *Manager) CloneSnapshot(ctx context.Context, snapshotName, cloneName string) error {
	if snapshotName == "" || cloneName == "" {
		return fmt.Errorf("snapshot name and clone name are required")
	}

	if !strings.Contains(snapshotName, "@") {
		return fmt.Errorf("invalid snapshot name format (expected dataset@snapshot)")
	}

	snapshot, err := gozfs.GetDataset(snapshotName)
	if err != nil {
		return fmt.Errorf("snapshot not found: %s: %w", snapshotName, err)
	}

	_, err = snapshot.Clone(cloneName, nil)
	if err != nil {
		return fmt.Errorf("failed to clone snapshot: %w", err)
	}

	return nil
}

// parseZFSTimestamp parses ZFS creation timestamp (Unix epoch as string).
func parseZFSTimestamp(timestamp string) (time.Time, error) {
	var epoch int64
	_, err := fmt.Sscanf(timestamp, "%d", &epoch)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(epoch, 0), nil
}
