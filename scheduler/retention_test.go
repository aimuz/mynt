package scheduler

import (
	"testing"
	"time"
)

func TestParseRetention(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "hours",
			input:   "24h",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "days",
			input:   "7d",
			want:    7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "30_days",
			input:   "30d",
			want:    30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "uppercase",
			input:   "24H",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "with_spaces",
			input:   "  7d  ",
			want:    7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "forever",
			input:   "forever",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid_format",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid_unit",
			input:   "7w",
			want:    0,
			wantErr: true,
		},
		{
			name:    "no_unit",
			input:   "7",
			want:    0,
			wantErr: true,
		},
		{
			name:    "no_number",
			input:   "h",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRetention(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRetention() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseRetention() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSnapshotTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid",
			input:   "20231208-143022",
			wantErr: false,
		},
		{
			name:    "valid_midnight",
			input:   "20231208-000000",
			wantErr: false,
		},
		{
			name:    "valid_end_of_day",
			input:   "20231208-235959",
			wantErr: false,
		},
		{
			name:    "invalid_format",
			input:   "2023-12-08 14:30:22",
			wantErr: true,
		},
		{
			name:    "invalid_date",
			input:   "20231301-143022",
			wantErr: true,
		},
		{
			name:    "invalid_time",
			input:   "20231208-256022",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSnapshotTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSnapshotTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify we can format it back
				formatted := got.Format("20060102-150405")
				if formatted != tt.input {
					t.Errorf("parseSnapshotTimestamp() round-trip failed: got %v, want %v", formatted, tt.input)
				}
			}
		})
	}
}

func TestParseSnapshotTimestamp_Values(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantHour  int
		wantMin   int
		wantSec   int
	}{
		{
			name:      "specific_datetime",
			input:     "20231208-143022",
			wantYear:  2023,
			wantMonth: time.December,
			wantDay:   8,
			wantHour:  14,
			wantMin:   30,
			wantSec:   22,
		},
		{
			name:      "midnight",
			input:     "20240101-000000",
			wantYear:  2024,
			wantMonth: time.January,
			wantDay:   1,
			wantHour:  0,
			wantMin:   0,
			wantSec:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSnapshotTimestamp(tt.input)
			if err != nil {
				t.Fatalf("parseSnapshotTimestamp() error = %v", err)
			}
			if got.Year() != tt.wantYear {
				t.Errorf("Year = %v, want %v", got.Year(), tt.wantYear)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("Month = %v, want %v", got.Month(), tt.wantMonth)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("Day = %v, want %v", got.Day(), tt.wantDay)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %v, want %v", got.Hour(), tt.wantHour)
			}
			if got.Minute() != tt.wantMin {
				t.Errorf("Minute = %v, want %v", got.Minute(), tt.wantMin)
			}
			if got.Second() != tt.wantSec {
				t.Errorf("Second = %v, want %v", got.Second(), tt.wantSec)
			}
		})
	}
}
