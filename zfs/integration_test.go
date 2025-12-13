package zfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.aimuz.me/mynt/testutil"
)

// testPoolName is the pool name used for integration tests.
const testPoolName = "testpool"

// setupTestPool creates a temporary file-backed ZFS pool for testing.
// Returns a cleanup function that destroys the pool and removes the backing file.
func setupTestPool(t *testing.T) *Manager {
	t.Helper()

	m := NewManager()
	ctx := context.Background()

	// Create a temporary file to use as a vdev
	tmpDir := t.TempDir()
	vdevPath := filepath.Join(tmpDir, "vdev.img")

	// Create a 100MB sparse file
	f, err := os.Create(vdevPath)
	if err != nil {
		t.Fatalf("failed to create vdev file: %v", err)
	}
	if err := f.Truncate(100 * 1024 * 1024); err != nil {
		f.Close()
		t.Fatalf("failed to truncate vdev file: %v", err)
	}
	f.Close()

	// Create the pool
	err = m.CreatePool(ctx, CreatePoolRequest{
		Name:    testPoolName,
		Devices: []string{vdevPath},
		Type:    "", // stripe
	})
	if err != nil {
		t.Fatalf("failed to create test pool: %v", err)
	}
	t.Cleanup(func() {
		_ = m.DestroyPool(ctx, testPoolName)
		os.Remove(vdevPath)
	})

	return m
}

func TestIntegration_Pool(t *testing.T) {
	testutil.RequireIntegration(t)

	m := setupTestPool(t)

	ctx := context.Background()

	t.Run("ListPools", func(t *testing.T) {
		pools, err := m.ListPools(ctx)
		if err != nil {
			t.Fatalf("ListPools: %v", err)
		}

		var found bool
		for _, p := range pools {
			if p.Name == testPoolName {
				found = true
				if p.Health != PoolOnline {
					t.Errorf("pool health = %v, want %v", p.Health, PoolOnline)
				}
				break
			}
		}
		if !found {
			t.Errorf("pool %q not found", testPoolName)
		}
	})

	t.Run("GetPool", func(t *testing.T) {
		pool, err := m.GetPool(ctx, testPoolName)
		if err != nil {
			t.Fatalf("GetPool: %v", err)
		}
		if pool.Name != testPoolName {
			t.Errorf("pool.Name = %q, want %q", pool.Name, testPoolName)
		}
	})
}

func TestIntegration_Dataset(t *testing.T) {
	testutil.RequireIntegration(t)

	m := setupTestPool(t)

	ctx := context.Background()
	datasetName := testPoolName + "/testdata"

	t.Run("Create", func(t *testing.T) {
		err := m.CreateDataset(ctx, CreateDatasetRequest{
			Name:    datasetName,
			Type:    "filesystem",
			UseCase: UseCaseGeneral,
		})
		if err != nil {
			t.Fatalf("CreateDataset: %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		ds, err := m.GetDataset(ctx, datasetName)
		if err != nil {
			t.Fatalf("GetDataset: %v", err)
		}
		if ds.Name != datasetName {
			t.Errorf("Name = %q, want %q", ds.Name, datasetName)
		}
		if ds.Type != DatasetFilesystem {
			t.Errorf("Type = %v, want %v", ds.Type, DatasetFilesystem)
		}
	})

	t.Run("SetProperty", func(t *testing.T) {
		if err := m.SetProperty(ctx, datasetName, "compression", "zstd"); err != nil {
			t.Errorf("SetProperty: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		datasets, err := m.ListDatasets(ctx)
		if err != nil {
			t.Fatalf("ListDatasets: %v", err)
		}

		var found bool
		for _, d := range datasets {
			if d.Name == datasetName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dataset %q not found in list", datasetName)
		}
	})

	t.Run("Destroy", func(t *testing.T) {
		if err := m.DestroyDataset(ctx, datasetName); err != nil {
			t.Fatalf("DestroyDataset: %v", err)
		}

		// Verify it's gone
		if _, err := m.GetDataset(ctx, datasetName); err == nil {
			t.Error("expected error after destroying dataset")
		}
	})
}

func TestIntegration_Snapshot(t *testing.T) {
	testutil.RequireIntegration(t)

	m := setupTestPool(t)

	ctx := context.Background()
	datasetName := testPoolName + "/snaptest"
	snapshotFullName := fmt.Sprintf("%s@snap1", datasetName)

	// Create a dataset first
	if err := m.CreateDataset(ctx, CreateDatasetRequest{
		Name: datasetName,
		Type: "filesystem",
	}); err != nil {
		t.Fatalf("CreateDataset: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		snap, err := m.CreateSnapshot(ctx, CreateSnapshotRequest{
			Dataset: datasetName,
			Name:    "snap1",
		})
		if err != nil {
			t.Fatalf("CreateSnapshot: %v", err)
		}
		if snap.Name != snapshotFullName {
			t.Errorf("Name = %q, want %q", snap.Name, snapshotFullName)
		}
		if snap.Source != "manual" {
			t.Errorf("Source = %q, want %q", snap.Source, "manual")
		}
	})

	t.Run("List", func(t *testing.T) {
		snapshots, err := m.ListSnapshots(ctx, datasetName)
		if err != nil {
			t.Fatalf("ListSnapshots: %v", err)
		}
		if len(snapshots) != 1 {
			t.Fatalf("len(snapshots) = %d, want 1", len(snapshots))
		}

		snap := snapshots[0]

		// Verify Name is parsed correctly
		if snap.Name != snapshotFullName {
			t.Errorf("Name = %q, want %q", snap.Name, snapshotFullName)
		}

		// Verify Dataset is set correctly
		if snap.Dataset != datasetName {
			t.Errorf("Dataset = %q, want %q", snap.Dataset, datasetName)
		}

		// Verify CreatedAt is parsed (should be RFC3339 format)
		if snap.CreatedAt == "" {
			t.Error("CreatedAt is empty, expected RFC3339 timestamp")
		}

		// Verify Source detection works
		if snap.Source != "manual" {
			t.Errorf("Source = %q, want %q", snap.Source, "manual")
		}

		// Used and Referenced should be non-negative (0 is valid for empty snapshot)
		// Just log them for visibility
		t.Logf("Snapshot fields: Name=%s, Dataset=%s, CreatedAt=%s, Used=%d, Referenced=%d, Source=%s",
			snap.Name, snap.Dataset, snap.CreatedAt, snap.Used, snap.Referenced, snap.Source)
	})

	t.Run("Destroy", func(t *testing.T) {
		if err := m.DestroySnapshot(ctx, snapshotFullName); err != nil {
			t.Fatalf("DestroySnapshot: %v", err)
		}

		// Verify it's gone
		snapshots, err := m.ListSnapshots(ctx, datasetName)
		if err != nil {
			t.Fatalf("ListSnapshots after destroy: %v", err)
		}
		if len(snapshots) != 0 {
			t.Errorf("expected 0 snapshots, got %d", len(snapshots))
		}
	})
}

func TestIntegration_Volume(t *testing.T) {
	testutil.RequireIntegration(t)

	m := setupTestPool(t)

	ctx := context.Background()
	volumeName := testPoolName + "/testvol"

	t.Run("Create", func(t *testing.T) {
		err := m.CreateDataset(ctx, CreateDatasetRequest{
			Name:  volumeName,
			Type:  "volume",
			Quota: 10 * 1024 * 1024, // 10MB
		})
		if err != nil {
			t.Fatalf("CreateDataset (volume): %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		ds, err := m.GetDataset(ctx, volumeName)
		if err != nil {
			t.Fatalf("GetDataset (volume): %v", err)
		}
		if ds.Type != DatasetVolume {
			t.Errorf("Type = %v, want %v", ds.Type, DatasetVolume)
		}
	})

	t.Run("Destroy", func(t *testing.T) {
		if err := m.DestroyDataset(ctx, volumeName); err != nil {
			t.Fatalf("DestroyDataset (volume): %v", err)
		}
	})
}

// TestIntegration_ListSnapshots_MultipleSnapshots verifies that ListSnapshots
// correctly parses multiple snapshots with different naming patterns.
func TestIntegration_ListSnapshots_MultipleSnapshots(t *testing.T) {
	testutil.RequireIntegration(t)

	m := setupTestPool(t)

	ctx := context.Background()
	datasetName := testPoolName + "/multisnap"

	// Create dataset
	if err := m.CreateDataset(ctx, CreateDatasetRequest{
		Name: datasetName,
		Type: "filesystem",
	}); err != nil {
		t.Fatalf("CreateDataset: %v", err)
	}

	// Create multiple snapshots with different naming patterns
	snapshotNames := []struct {
		name       string
		wantSource string
	}{
		{"manual-backup", "manual"},
		{"auto-daily-20241213-120000", "policy:daily"},
		{"snap2", "manual"},
	}

	for _, sn := range snapshotNames {
		_, err := m.CreateSnapshot(ctx, CreateSnapshotRequest{
			Dataset: datasetName,
			Name:    sn.name,
		})
		if err != nil {
			t.Fatalf("CreateSnapshot(%s): %v", sn.name, err)
		}
	}

	// List and verify
	snapshots, err := m.ListSnapshots(ctx, datasetName)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}

	if len(snapshots) != len(snapshotNames) {
		t.Fatalf("len(snapshots) = %d, want %d", len(snapshots), len(snapshotNames))
	}

	// Build a map for easier lookup
	snapByName := make(map[string]Snapshot)
	for _, s := range snapshots {
		snapByName[s.Name] = s
	}

	// Verify each snapshot's fields
	for _, sn := range snapshotNames {
		fullName := fmt.Sprintf("%s@%s", datasetName, sn.name)
		snap, ok := snapByName[fullName]
		if !ok {
			t.Errorf("snapshot %q not found in list", fullName)
			continue
		}

		if snap.Dataset != datasetName {
			t.Errorf("%s: Dataset = %q, want %q", sn.name, snap.Dataset, datasetName)
		}
		if snap.Source != sn.wantSource {
			t.Errorf("%s: Source = %q, want %q", sn.name, snap.Source, sn.wantSource)
		}
		if snap.CreatedAt == "" {
			t.Errorf("%s: CreatedAt is empty", sn.name)
		}

		t.Logf("Snapshot %s: CreatedAt=%s, Used=%d, Referenced=%d, Source=%s",
			sn.name, snap.CreatedAt, snap.Used, snap.Referenced, snap.Source)
	}

	// Verify snapshots are sorted by CreatedAt
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i-1].CreatedAt > snapshots[i].CreatedAt {
			t.Errorf("snapshots not sorted by CreatedAt: %s > %s",
				snapshots[i-1].CreatedAt, snapshots[i].CreatedAt)
		}
	}
}
