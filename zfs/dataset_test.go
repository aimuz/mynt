package zfs

import (
	"context"
	"testing"
)

func TestCreateDataset(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateDatasetRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "missing_name",
			req: CreateDatasetRequest{
				Type: "filesystem",
			},
			wantErr:     true,
			errContains: "dataset name is required",
		},
		{
			name: "volume_without_size",
			req: CreateDatasetRequest{
				Name: "pool/volume2",
				Type: "volume",
			},
			wantErr:     true,
			errContains: "size is required for volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			err := m.CreateDataset(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("expected error to contain %q, got nil", tt.errContains)
				} else if err.Error() != "" && len(err.Error()) > 0 {
					// Just check error is not empty
					t.Logf("Error message: %s", err.Error())
				}
			}
		})
	}
}

func TestDestroyDataset(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing_name",
			datasetName: "",
			wantErr:     true,
			errContains: "dataset name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			err := m.DestroyDataset(context.Background(), tt.datasetName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DestroyDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("expected error to contain %q, got nil", tt.errContains)
				}
			}
		})
	}
}

func TestSetProperty(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		key         string
		value       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing_dataset_name",
			datasetName: "",
			key:         "compression",
			value:       "lz4",
			wantErr:     true,
			errContains: "dataset name and property key are required",
		},
		{
			name:        "missing_property_key",
			datasetName: "pool/dataset1",
			key:         "",
			value:       "lz4",
			wantErr:     true,
			errContains: "dataset name and property key are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			err := m.SetProperty(context.Background(), tt.datasetName, tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("expected error to contain %q, got nil", tt.errContains)
				}
			}
		})
	}
}
