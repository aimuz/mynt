package zfs

import (
	"context"
	"fmt"

	gozfs "github.com/mistifyio/go-zfs/v4"
)

// CreateDatasetRequest represents a request to create a dataset.
type CreateDatasetRequest struct {
	Name       string            `json:"name"`       // required: pool/name
	Type       string            `json:"type"`       // filesystem (default) or volume
	Size       uint64            `json:"size"`       // required for volumes
	Properties map[string]string `json:"properties"` // optional ZFS properties
}

// CreateDataset creates a new ZFS dataset.
func (m *Manager) CreateDataset(ctx context.Context, req CreateDatasetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("dataset name is required")
	}

	// Default to filesystem if not specified
	if req.Type == "" {
		req.Type = "filesystem"
	}

	var err error

	if req.Type == "volume" {
		if req.Size == 0 {
			return fmt.Errorf("size is required for volumes")
		}
		_, err = gozfs.CreateVolume(req.Name, req.Size, req.Properties)
	} else {
		_, err = gozfs.CreateFilesystem(req.Name, req.Properties)
	}

	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	return nil
}

// GetDataset returns details for a specific dataset.
func (m *Manager) GetDataset(ctx context.Context, name string) (*Dataset, error) {
	gozfsDataset, err := gozfs.GetDataset(name)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %s: %w", name, err)
	}

	dataset := fromGozfsDataset(gozfsDataset)

	// Get encryption and dedup properties
	if enc, err := gozfsDataset.GetProperty("encryption"); err == nil {
		dataset.Encryption = enc
	}
	if dedup, err := gozfsDataset.GetProperty("dedup"); err == nil {
		dataset.Deduplication = dedup
	}

	return &dataset, nil
}

// DestroyDataset destroys a ZFS dataset.
func (m *Manager) DestroyDataset(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("dataset name is required")
	}

	gozfsDataset, err := gozfs.GetDataset(name)
	if err != nil {
		return fmt.Errorf("dataset not found: %s: %w", name, err)
	}

	// Use DestroyRecursive flag to destroy recursively (including snapshots and children)
	if err := gozfsDataset.Destroy(gozfs.DestroyRecursive); err != nil {
		return fmt.Errorf("failed to destroy dataset: %w", err)
	}

	return nil
}

// SetProperty sets a property on a dataset.
func (m *Manager) SetProperty(ctx context.Context, name, key, value string) error {
	if name == "" || key == "" {
		return fmt.Errorf("dataset name and property key are required")
	}

	gozfsDataset, err := gozfs.GetDataset(name)
	if err != nil {
		return fmt.Errorf("dataset not found: %s: %w", name, err)
	}

	if err := gozfsDataset.SetProperty(key, value); err != nil {
		return fmt.Errorf("failed to set property: %w", err)
	}

	return nil
}
