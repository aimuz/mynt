package zfs

import (
	"context"
	"strings"
	"testing"
)

func TestParseVDevsFromJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]*Vdev
		wantCount int
		wantTypes []string
	}{
		{
			name:      "empty",
			input:     nil,
			wantCount: 0,
		},
		{
			name: "stripe_pool",
			input: map[string]*Vdev{
				"test": {
					Name:     "test",
					VDevType: "root",
					State:    "ONLINE",
					VDevs: map[string]*Vdev{
						"sda": {
							Name:     "sda",
							VDevType: "disk",
							Path:     "/dev/sda",
							State:    "ONLINE",
						},
						"sdb": {
							Name:     "sdb",
							VDevType: "disk",
							Path:     "/dev/sdb",
							State:    "ONLINE",
						},
					},
				},
			},
			wantCount: 2, // Two stripe disks become two vdevs
			wantTypes: []string{"stripe", "stripe"},
		},
		{
			name: "mirror_pool",
			input: map[string]*Vdev{
				"tank": {
					Name:     "tank",
					VDevType: "root",
					State:    "ONLINE",
					VDevs: map[string]*Vdev{
						"mirror-0": {
							Name:     "mirror-0",
							VDevType: "mirror",
							State:    "ONLINE",
							VDevs: map[string]*Vdev{
								"sda": {
									Name:     "sda",
									VDevType: "disk",
									Path:     "/dev/sda",
									State:    "ONLINE",
								},
								"sdb": {
									Name:     "sdb",
									VDevType: "disk",
									Path:     "/dev/sdb",
									State:    "ONLINE",
								},
							},
						},
					},
				},
			},
			wantCount: 1, // One mirror vdev
			wantTypes: []string{"mirror"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVDevsFromJSON(tt.input)

			if len(got) != tt.wantCount {
				t.Errorf("parseVDevsFromJSON() count = %d, want %d", len(got), tt.wantCount)
			}

			if tt.wantTypes != nil {
				for i, vdev := range got {
					if i < len(tt.wantTypes) && vdev.Type != tt.wantTypes[i] {
						t.Errorf("vdev[%d].Type = %q, want %q", i, vdev.Type, tt.wantTypes[i])
					}
				}
			}
		})
	}
}

func TestParseResilverFromJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        *ScanStatsJSON
		wantProgress bool
		wantPercent  float64
	}{
		{
			name:         "nil_scan",
			input:        nil,
			wantProgress: false,
		},
		{
			name: "scrub_finished",
			input: &ScanStatsJSON{
				Function: "SCRUB",
				State:    "FINISHED",
			},
			wantProgress: false,
		},
		{
			name: "resilver_in_progress",
			input: &ScanStatsJSON{
				Function:     "RESILVER",
				State:        "SCANNING",
				PassStart:    "1733648000",    // Unix timestamp
				Examined:     "536870912000",  // 500G in bytes
				Issued:       "536870912000",  // 500G issued
				ToExamine:    "1099511627776", // 1T in bytes
				BytesPerScan: "104857600",     // 100M in bytes
			},
			wantProgress: true,
			wantPercent:  48.83, // 500G / 1T ≈ 48.83%
		},
		{
			name: "resilver_finished",
			input: &ScanStatsJSON{
				Function: "RESILVER",
				State:    "FINISHED",
			},
			wantProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResilverFromJSON(tt.input)

			if got.InProgress != tt.wantProgress {
				t.Errorf("InProgress = %v, want %v", got.InProgress, tt.wantProgress)
			}

			if tt.wantProgress {
				// Allow 1% tolerance for floating point
				if got.PercentDone < tt.wantPercent-1 || got.PercentDone > tt.wantPercent+1 {
					t.Errorf("PercentDone = %v, want ~%v", got.PercentDone, tt.wantPercent)
				}
			}
		})
	}
}

func TestVdevTypeFromJSON(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mirror", "mirror"},
		{"raidz", "raidz"},
		{"raidz1", "raidz"},
		{"raidz2", "raidz2"},
		{"raidz3", "raidz3"},
		{"disk", "stripe"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := vdevTypeFromJSON(tt.input)
			if got != tt.want {
				t.Errorf("vdevTypeFromJSON(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildDataset_MissingProperties(t *testing.T) {
	// this dataset is missing all properties
	dj := &DatasetListJSON{
		Name: "pool/test",
		Type: "FILESYSTEM",
		Pool: "pool",
	}

	// This should not panic
	ds := buildDataset(dj)

	if ds.Name != "pool/test" {
		t.Errorf("expected name 'pool/test', got '%s'", ds.Name)
	}
	if ds.Used != 0 {
		t.Errorf("expected used 0, got %d", ds.Used)
	}
	if ds.Available != 0 {
		t.Errorf("expected available 0, got %d", ds.Available)
	}
}

func TestListDatasets_Validation(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		wantErr     bool
	}{
		{
			name:        "valid_simple",
			datasetName: "pool/test",
			wantErr:     false,
		},
		{
			name:        "valid_with_snapshot",
			datasetName: "pool/test@snap",
			wantErr:     false,
		},
		{
			name:        "valid_chars",
			datasetName: "pool/a-b_c.d:e",
			wantErr:     false,
		},
		{
			name:        "invalid_semicolon",
			datasetName: "pool/test;rm",
			wantErr:     true,
		},
		{
			name:        "invalid_pipe",
			datasetName: "pool/test|ls",
			wantErr:     true,
		},
		{
			name:        "invalid_backtick",
			datasetName: "pool/test`",
			wantErr:     true,
		},
		{
			name:        "invalid_dollar",
			datasetName: "$(whoami)",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			// We only care about validation error, ignore execution errors
			// since we don't have a real ZFS environment
			_, err := m.listDatasets(context.Background(), tt.datasetName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.datasetName)
				}
			} else {
				// If we expect success (validation pass), we might still get "zfs list" error
				// because command fails, but that's fine. We check if error is NOT validation error.
				if err != nil && strings.Contains(err.Error(), "invalid character") {
					t.Errorf("unexpected validation error for input %q: %v", tt.datasetName, err)
				}
			}
		})
	}
}

func TestListPools_Validation(t *testing.T) {
	tests := []struct {
		name     string
		poolName string
		wantErr  bool
	}{
		{
			name:     "valid_simple",
			poolName: "tank",
			wantErr:  false,
		},
		{
			name:     "invalid_semicolon",
			poolName: "tank;rm",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			// We only care about validation error, ignore execution errors
			// Use internal listPools which accepts varargs
			_, err := m.listPools(context.Background(), tt.poolName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.poolName)
				}
			} else {
				if err != nil && strings.Contains(err.Error(), "invalid character") {
					t.Errorf("unexpected validation error for input %q: %v", tt.poolName, err)
				}
			}
		})
	}
}
