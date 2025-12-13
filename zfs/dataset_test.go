package zfs

import (
	"context"
	"strings"
	"testing"
)

func TestCreateDataset_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateDatasetRequest
		wantErr string
	}{
		{
			name:    "missing_name",
			req:     CreateDatasetRequest{Type: "filesystem"},
			wantErr: "dataset name is required",
		},
		{
			name:    "volume_without_size",
			req:     CreateDatasetRequest{Name: "pool/volume2", Type: "volume"},
			wantErr: "quota (size) is required for volumes",
		},
	}

	m := NewManager()
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.CreateDataset(ctx, tt.req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDestroyDataset_Validation(t *testing.T) {
	m := NewManager()
	err := m.DestroyDataset(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty dataset name")
	}
	if want := "dataset name is required"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want containing %q", err.Error(), want)
	}
}

func TestSetProperty_Validation(t *testing.T) {
	tests := []struct {
		name    string
		dataset string
		key     string
		value   string
		wantErr string
	}{
		{
			name:    "missing_dataset",
			dataset: "",
			key:     "compression",
			value:   "lz4",
			wantErr: "dataset name and property key are required",
		},
		{
			name:    "missing_key",
			dataset: "pool/dataset1",
			key:     "",
			value:   "lz4",
			wantErr: "dataset name and property key are required",
		},
	}

	m := NewManager()
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.SetProperty(ctx, tt.dataset, tt.key, tt.value)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}
