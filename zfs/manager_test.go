package zfs

import (
	"testing"
)

func TestParseResilverStatus(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProgress bool
		wantPercent  float64
		wantScanned  uint64
		wantTotal    uint64
		wantRate     uint64
		wantTime     string
	}{
		{
			name:         "no_resilver",
			input:        "  pool: tank\n state: ONLINE\n",
			wantProgress: false,
		},
		{
			name: "resilver_in_progress",
			input: `  pool: tank
 state: DEGRADED
status: One or more devices is currently being resilvered.
  scan: resilver in progress since Sun Dec  8 10:00:00 2024
        112G scanned of 1.81T at 1.14G/s, 0h24m to go
        50.0% done
`,
			wantProgress: true,
			wantPercent:  50.0,
			wantScanned:  120259084288,  // 112G
			wantTotal:    1990116046274, // 1.81T
			wantRate:     1224440233,    // 1.14G
			wantTime:     "0h24m",
		},
		{
			name: "resilver_almost_done",
			input: `  scan: resilver in progress since Sun Dec  8 10:00:00 2024
        1.5T scanned of 2T at 500M/s, 0:10:00 to go
        95.5% done
`,
			wantProgress: true,
			wantPercent:  95.5,
			wantScanned:  1649267441664, // 1.5T
			wantTotal:    2199023255552, // 2T
			wantRate:     524288000,     // 500M
			wantTime:     "0:10:00",
		},
		{
			name: "scrub_not_resilver",
			input: `  scan: scrub in progress since Sun Dec  8 10:00:00 2024
        100G scanned at 200M/s
        50.0% done
`,
			wantProgress: false, // scrub, not resilver
			wantPercent:  50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResilverStatus(tt.input)

			if got.InProgress != tt.wantProgress {
				t.Errorf("InProgress = %v, want %v", got.InProgress, tt.wantProgress)
			}

			if got.PercentDone != tt.wantPercent {
				t.Errorf("PercentDone = %v, want %v", got.PercentDone, tt.wantPercent)
			}

			if tt.wantProgress {
				// Allow 5% tolerance for parsed sizes due to floating point
				if !withinTolerance(got.ScannedBytes, tt.wantScanned, 0.05) {
					t.Errorf("ScannedBytes = %d, want ~%d", got.ScannedBytes, tt.wantScanned)
				}

				if !withinTolerance(got.TotalBytes, tt.wantTotal, 0.05) {
					t.Errorf("TotalBytes = %d, want ~%d", got.TotalBytes, tt.wantTotal)
				}

				if !withinTolerance(got.Rate, tt.wantRate, 0.05) {
					t.Errorf("Rate = %d, want ~%d", got.Rate, tt.wantRate)
				}

				if got.TimeRemaining != tt.wantTime {
					t.Errorf("TimeRemaining = %q, want %q", got.TimeRemaining, tt.wantTime)
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
		{"1.81T", 1990585565593},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSize(tt.input)
			if !withinTolerance(got, tt.want, 0.01) {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseScanLine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantScanned uint64
		wantTotal   uint64
		wantRate    uint64
	}{
		{
			name:        "standard_format",
			input:       "112G scanned of 1.81T at 1.14G/s, 0h24m to go",
			wantScanned: 120259084288,  // 112G
			wantTotal:   1990116046274, // 1.81T
			wantRate:    1224440233,    // 1.14G
		},
		{
			name:        "megabytes",
			input:       "500M scanned of 10G at 100M/s",
			wantScanned: 524288000,   // 500M
			wantTotal:   10737418240, // 10G
			wantRate:    104857600,   // 100M
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanned, total, rate := parseScanLine(tt.input)

			if !withinTolerance(scanned, tt.wantScanned, 0.05) {
				t.Errorf("scanned = %d, want ~%d", scanned, tt.wantScanned)
			}

			if !withinTolerance(total, tt.wantTotal, 0.05) {
				t.Errorf("total = %d, want ~%d", total, tt.wantTotal)
			}

			if !withinTolerance(rate, tt.wantRate, 0.05) {
				t.Errorf("rate = %d, want ~%d", rate, tt.wantRate)
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
