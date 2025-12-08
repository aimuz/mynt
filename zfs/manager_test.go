package zfs

import (
	"testing"
)

func TestParseVDevsFromJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]*VDevJSON
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
			input: map[string]*VDevJSON{
				"test": {
					Name:     "test",
					VDevType: "root",
					State:    "ONLINE",
					VDevs: map[string]*VDevJSON{
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
			input: map[string]*VDevJSON{
				"tank": {
					Name:     "tank",
					VDevType: "root",
					State:    "ONLINE",
					VDevs: map[string]*VDevJSON{
						"mirror-0": {
							Name:     "mirror-0",
							VDevType: "mirror",
							State:    "ONLINE",
							VDevs: map[string]*VDevJSON{
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
				Examined:     "500G",
				ToExamine:    "1T",
				BytesPerScan: "100M",
			},
			wantProgress: true,
			wantPercent:  48.83, // 500G / 1T = 500*1024^3 / 1024^4 ≈ 48.83%
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

func TestParseSize(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"100", 100},
		{"1K", 1024},
		{"1k", 1024},
		{"100M", 100 * 1024 * 1024},
		{"1G", 1024 * 1024 * 1024},
		{"2T", 2 * 1024 * 1024 * 1024 * 1024},
		{"1.5G", 1610612736}, // 1.5 * 1024^3
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSize(tt.input)
			// Allow 5% tolerance for floating point sizes
			if !withinTolerance(got, tt.want, 0.05) {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
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

// withinTolerance checks if got is within tolerance percent of want.
func withinTolerance(got, want uint64, tolerance float64) bool {
	if want == 0 {
		return got == 0
	}
	diff := float64(got) - float64(want)
	if diff < 0 {
		diff = -diff
	}
	return diff/float64(want) <= tolerance
}
