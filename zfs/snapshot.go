package zfs

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
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

const zfsSnapshotProperties = "name,used,referenced,creation"

// ListSnapshots returns all snapshots for a specific dataset.
func (m *Manager) ListSnapshots(ctx context.Context, datasetName string) ([]Snapshot, error) {
	if datasetName == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	args := []string{"list", "-j", "-p", "-t", "snapshot", "-o", zfsSnapshotProperties, datasetName}
	out, err := m.exec.Output(ctx, "zfs", args...)
	if err != nil {
		return nil, fmt.Errorf("zfs list snapshots: %w", err)
	}

	var listJSON ZFSListJSON
	if err := json.Unmarshal(out, &listJSON); err != nil {
		return nil, fmt.Errorf("parse zfs list snapshots: %w", err)
	}

	snapshots := make([]Snapshot, 0, len(listJSON.Datasets))
	for _, sj := range listJSON.Datasets {
		snapshots = append(snapshots, buildSnapshot(sj, datasetName))
	}

	slices.SortFunc(snapshots, func(a, b Snapshot) int {
		return cmp.Or(
			strings.Compare(a.CreatedAt, b.CreatedAt),
			strings.Compare(a.Name, b.Name),
		)
	})

	return snapshots, nil
}

// buildSnapshot constructs a Snapshot from JSON data.
func buildSnapshot(sj *DatasetListJSON, datasetName string) Snapshot {
	// Parse creation time from Unix epoch
	var createdAt string
	if prop := sj.Properties["creation"]; prop != nil {
		if t, err := parseZFSTimestamp(prop.Value); err == nil {
			createdAt = t.Format(time.RFC3339)
		}
	}

	return Snapshot{
		Name:       sj.Name,
		Dataset:    datasetName,
		CreatedAt:  createdAt,
		Used:       parseUint(sj.GetProp("used")),
		Referenced: parseUint(sj.GetProp("referenced")),
		Source:     detectSnapshotSource(sj.Name),
	}
}

const timestampSuffixLen = 16

// detectSnapshotSource determines if a snapshot was created manually or by policy.
func detectSnapshotSource(snapshotName string) string {
	parts := strings.Split(snapshotName, "@")
	if len(parts) != 2 {
		return "manual"
	}

	snapName := parts[1]
	if !strings.HasPrefix(snapName, "auto-") {
		return "manual"
	}

	// Format: auto-{policyName}-{timestamp} where timestamp is YYYYMMDD-HHMMSS (16 chars)
	rest := strings.TrimPrefix(snapName, "auto-")
	if len(rest) > timestampSuffixLen {
		return "policy:" + rest[:len(rest)-timestampSuffixLen]
	}
	return "policy:auto"
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
