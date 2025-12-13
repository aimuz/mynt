package zfs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		datasetName string
		wantCount   int
		checks      []struct {
			name       string
			wantSource string
		}
	}{
		{
			name:        "single_snapshot",
			file:        "list_snapshots_single.json",
			datasetName: "testpool/data",
			wantCount:   1,
			checks: []struct {
				name       string
				wantSource string
			}{
				{"testpool/data@snap1", "manual"},
			},
		},
		{
			name:        "multiple_snapshots",
			file:        "list_snapshots_multiple.json",
			datasetName: "testpool/data",
			wantCount:   3,
			checks: []struct {
				name       string
				wantSource string
			}{
				{"testpool/data@manual-backup", "manual"},
				{"testpool/data@auto-daily-20241213-120000", "policy:daily"},
				{"testpool/data@auto-hourly-20241214-080000", "policy:hourly"},
			},
		},
		{
			name:        "empty_list",
			file:        "list_snapshots_empty.json",
			datasetName: "testpool/data",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("read testdata: %v", err)
			}

			var listJSON ZFSListJSON
			if err := json.Unmarshal(data, &listJSON); err != nil {
				t.Fatalf("parse JSON: %v", err)
			}

			if len(listJSON.Datasets) != tt.wantCount {
				t.Fatalf("len(Datasets) = %d, want %d", len(listJSON.Datasets), tt.wantCount)
			}

			// Build snapshots and verify
			snapByName := make(map[string]Snapshot)
			for _, dj := range listJSON.Datasets {
				snap := buildSnapshot(dj, tt.datasetName)
				snapByName[snap.Name] = snap
			}

			for _, check := range tt.checks {
				snap, ok := snapByName[check.name]
				if !ok {
					t.Errorf("snapshot %q not found", check.name)
					continue
				}

				if snap.Dataset != tt.datasetName {
					t.Errorf("%s: Dataset = %q, want %q", check.name, snap.Dataset, tt.datasetName)
				}
				if snap.Source != check.wantSource {
					t.Errorf("%s: Source = %q, want %q", check.name, snap.Source, check.wantSource)
				}
				if snap.CreatedAt == "" {
					t.Errorf("%s: CreatedAt is empty", check.name)
				}
			}
		})
	}
}
