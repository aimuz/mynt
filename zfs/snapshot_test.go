package zfs

import (
	"strings"
	"testing"
)

func TestCreateSnapshot_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSnapshotRequest
		wantErr string
	}{
		{
			name:    "missing_dataset",
			req:     CreateSnapshotRequest{Name: "snap1"},
			wantErr: "dataset name is required",
		},
		{
			name:    "missing_snapshot_name",
			req:     CreateSnapshotRequest{Dataset: "pool/data"},
			wantErr: "snapshot name is required",
		},
	}

	m := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.CreateSnapshot(nil, tt.req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDestroySnapshot_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"empty", "", "snapshot name is required"},
		{"no_at_sign", "pool/data", "invalid snapshot name format"},
	}

	m := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.DestroySnapshot(nil, tt.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRollbackSnapshot_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no_at_sign", "pool/data"},
	}

	m := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.RollbackSnapshot(nil, tt.input); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCloneSnapshot_Validation(t *testing.T) {
	tests := []struct {
		name      string
		snapshot  string
		cloneName string
	}{
		{"missing_snapshot", "", "pool/clone"},
		{"missing_clone_name", "pool/data@snap1", ""},
		{"invalid_snapshot_format", "pool/data", "pool/clone"},
	}

	m := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.CloneSnapshot(nil, tt.snapshot, tt.cloneName); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDetectSnapshotSource(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// Manual snapshots
		{"pool/data@backup", "manual"},
		{"pool/data@snap1", "manual"},
		{"pool/data@2024-01-01", "manual"},

		// Auto snapshots with policy
		{"pool/data@auto-daily-20241213-120000", "policy:daily"},
		{"pool/data@auto-weekly-20241213-120000", "policy:weekly"},
		{"pool/data@auto-hourly-20241213-120000", "policy:hourly"},

		// Edge cases
		{"pool/data@auto-", "policy:auto"},
		{"invalid_no_at", "manual"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectSnapshotSource(tt.name); got != tt.want {
				t.Errorf("detectSnapshotSource(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseZFSTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		wantYear int
		wantErr  bool
	}{
		{"1702468800", 2023, false}, // 2023-12-13 12:00:00 UTC
		{"0", 1970, false},          // epoch
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseZFSTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseZFSTimestamp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got.Year() != tt.wantYear {
				t.Errorf("year = %d, want %d", got.Year(), tt.wantYear)
			}
		})
	}
}

func TestListSnapshots_EmptyDataset(t *testing.T) {
	m := NewManager()
	_, err := m.ListSnapshots(nil, "")
	if err == nil {
		t.Error("expected error for empty dataset name")
	}
}
