package zfs

import (
	"context"
	"strings"
	"testing"

	"go.aimuz.me/mynt/sysexec"
)

func TestGetTemplateProperties(t *testing.T) {
	tests := []struct {
		useCase UseCaseTemplate
		wantMap map[string]string
	}{
		{
			UseCaseGeneral,
			map[string]string{"compression": "lz4", "recordsize": "128K"},
		},
		{
			UseCaseMedia,
			map[string]string{"recordsize": "1M", "atime": "off"},
		},
		{
			UseCaseDatabase,
			map[string]string{"recordsize": "16K", "logbias": "latency", "sync": "always"},
		},
		{
			UseCaseVM,
			map[string]string{"recordsize": "64K", "sync": "disabled"},
		},
		{
			UseCaseSurveillance,
			map[string]string{"recordsize": "1M"},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.useCase), func(t *testing.T) {
			got := GetTemplateProperties(tt.useCase)
			for k, want := range tt.wantMap {
				if got[k] != want {
					t.Errorf("GetTemplateProperties(%q)[%q] = %q, want %q", tt.useCase, k, got[k], want)
				}
			}
		})
	}
}

func TestParseUint(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"0", 0},
		{"1234567890", 1234567890},
		{"18446744073709551615", 18446744073709551615}, // max uint64
		{"", 0},
		{"invalid", 0},
		{"-1", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseUint(tt.input); got != tt.want {
				t.Errorf("parseUint(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestCalculateRedundancy(t *testing.T) {
	tests := []struct {
		name  string
		vdevs []VDevDetail
		want  int
	}{
		{"empty", nil, 0},
		{
			"stripe_no_redundancy",
			[]VDevDetail{
				{Type: "stripe", Status: "ONLINE"},
				{Type: "stripe", Status: "ONLINE"},
			},
			0,
		},
		{
			"mirror_healthy",
			[]VDevDetail{{
				Type:     "mirror",
				Status:   "ONLINE",
				Children: []DiskDetail{{Status: "ONLINE"}, {Status: "ONLINE"}},
			}},
			1,
		},
		{
			"mirror_degraded",
			[]VDevDetail{{
				Type:     "mirror",
				Status:   "DEGRADED",
				Children: []DiskDetail{{Status: "ONLINE"}, {Status: "FAULTED"}},
			}},
			0,
		},
		{
			"raidz1_healthy",
			[]VDevDetail{{
				Type:     "raidz",
				Status:   "ONLINE",
				Children: []DiskDetail{{Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}},
			}},
			1,
		},
		{
			"raidz2_healthy",
			[]VDevDetail{{
				Type:     "raidz2",
				Status:   "ONLINE",
				Children: []DiskDetail{{Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}},
			}},
			2,
		},
		{
			"raidz3_healthy",
			[]VDevDetail{{
				Type:     "raidz3",
				Status:   "ONLINE",
				Children: []DiskDetail{{Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}, {Status: "ONLINE"}},
			}},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateRedundancy(tt.vdevs); got != tt.want {
				t.Errorf("calculateRedundancy() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseScrubFromJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      *ScanStatsJSON
		wantNil    bool
		wantActive bool
	}{
		{"nil", nil, true, false},
		{
			"finished",
			&ScanStatsJSON{Function: "SCRUB", State: "FINISHED"},
			false, false,
		},
		{
			"in_progress",
			&ScanStatsJSON{Function: "SCRUB", State: "SCANNING", Examined: "1073741824", ToExamine: "10737418240"},
			false, true,
		},
		{
			"resilver_ignored",
			&ScanStatsJSON{Function: "RESILVER", State: "SCANNING"},
			true, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseScrubFromJSON(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("got %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("got nil, want non-nil")
			}
			if got.InProgress != tt.wantActive {
				t.Errorf("InProgress = %v, want %v", got.InProgress, tt.wantActive)
			}
		})
	}
}

func TestParseVDevsFromJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]*Vdev
		wantCount int
		wantTypes []string
	}{
		{"empty", nil, 0, nil},
		{
			"stripe_pool",
			map[string]*Vdev{
				"test": {
					Name: "test", VDevType: "root", State: "ONLINE",
					VDevs: map[string]*Vdev{
						"sda": {Name: "sda", VDevType: "disk", Path: "/dev/sda", State: "ONLINE"},
						"sdb": {Name: "sdb", VDevType: "disk", Path: "/dev/sdb", State: "ONLINE"},
					},
				},
			},
			2, []string{"stripe", "stripe"},
		},
		{
			"mirror_pool",
			map[string]*Vdev{
				"tank": {
					Name: "tank", VDevType: "root", State: "ONLINE",
					VDevs: map[string]*Vdev{
						"mirror-0": {
							Name: "mirror-0", VDevType: "mirror", State: "ONLINE",
							VDevs: map[string]*Vdev{
								"sda": {Name: "sda", VDevType: "disk", Path: "/dev/sda", State: "ONLINE"},
								"sdb": {Name: "sdb", VDevType: "disk", Path: "/dev/sdb", State: "ONLINE"},
							},
						},
					},
				},
			},
			1, []string{"mirror"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVDevsFromJSON(tt.input)
			if len(got) != tt.wantCount {
				t.Errorf("len(result) = %d, want %d", len(got), tt.wantCount)
			}
			for i, vdev := range got {
				if i < len(tt.wantTypes) && vdev.Type != tt.wantTypes[i] {
					t.Errorf("vdev[%d].Type = %q, want %q", i, vdev.Type, tt.wantTypes[i])
				}
			}
		})
	}
}

func TestParseResilverFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       *ScanStatsJSON
		wantActive  bool
		wantPercent float64
	}{
		{"nil", nil, false, 0},
		{
			"scrub_finished",
			&ScanStatsJSON{Function: "SCRUB", State: "FINISHED"},
			false, 0,
		},
		{
			"resilver_in_progress",
			&ScanStatsJSON{
				Function:     "RESILVER",
				State:        "SCANNING",
				PassStart:    "1733648000",
				Examined:     "536870912000", // 500G
				Issued:       "536870912000",
				ToExamine:    "1099511627776", // 1T
				BytesPerScan: "104857600",
			},
			true, 48.83,
		},
		{
			"resilver_finished",
			&ScanStatsJSON{Function: "RESILVER", State: "FINISHED"},
			false, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResilverFromJSON(tt.input)
			if got.InProgress != tt.wantActive {
				t.Errorf("InProgress = %v, want %v", got.InProgress, tt.wantActive)
			}
			if tt.wantActive {
				// 1% tolerance for floating point
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
			if got := vdevTypeFromJSON(tt.input); got != tt.want {
				t.Errorf("vdevTypeFromJSON(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildDataset_MissingProperties(t *testing.T) {
	// Dataset with nil Properties map.
	dj := &DatasetListJSON{
		Name: "pool/test",
		Type: "FILESYSTEM",
		Pool: "pool",
	}
	ds := buildDataset(dj)

	if ds.Name != "pool/test" {
		t.Errorf("Name = %q, want %q", ds.Name, "pool/test")
	}
	if ds.Used != 0 {
		t.Errorf("Used = %d, want 0", ds.Used)
	}
	if ds.Available != 0 {
		t.Errorf("Available = %d, want 0", ds.Available)
	}
}

func TestListDatasets_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid_simple", "pool/test", false},
		{"valid_with_snapshot", "pool/test@snap", false},
		{"valid_chars", "pool/a-b_c.d:e", false},
		{"invalid_semicolon", "pool/test;rm", true},
		{"invalid_pipe", "pool/test|ls", true},
		{"invalid_backtick", "pool/test`", true},
		{"invalid_dollar", "$(whoami)", true},
	}

	exec := sysexec.NewMock()
	exec.SetOutput("zfs", []byte(`{"output_version":{},"datasets":{}}`))
	m := &Manager{exec: exec}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.listDatasets(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
			} else {
				// Validation should pass; command may still fail.
				if err != nil && strings.Contains(err.Error(), "invalid character") {
					t.Errorf("unexpected validation error for %q: %v", tt.input, err)
				}
			}
		})
	}
}

func TestListPools_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid_simple", "tank", false},
		{"invalid_semicolon", "tank;rm", true},
	}

	exec := sysexec.NewMock()
	exec.SetOutput("zpool", []byte(`{"output_version":{},"pools":{}}`))
	m := &Manager{exec: exec}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.listPools(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
			} else {
				if err != nil && strings.Contains(err.Error(), "invalid character") {
					t.Errorf("unexpected validation error for %q: %v", tt.input, err)
				}
			}
		})
	}
}
