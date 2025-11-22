package zfs

import (
	"context"
	"fmt"
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

	args := []string{"create"}

	// Add properties
	for key, value := range req.Properties {
		args = append(args, "-o", fmt.Sprintf("%s=%s", key, value))
	}

	// For volumes, specify size with -V
	if req.Type == "volume" {
		if req.Size == 0 {
			return fmt.Errorf("size is required for volumes")
		}
		args = append(args, "-V", formatSize(req.Size))
	}

	args = append(args, req.Name)

	_, err := m.exec.Output(ctx, "zfs", args...)
	return err
}

// GetDataset returns details for a specific dataset.
func (m *Manager) GetDataset(ctx context.Context, name string) (*Dataset, error) {
	// For now, list all and find the one
	datasets, err := m.ListDatasets(ctx)
	if err != nil {
		return nil, err
	}

	for _, d := range datasets {
		if d.Name == name {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("dataset not found: %s", name)
}

// DestroyDataset destroys a ZFS dataset.
func (m *Manager) DestroyDataset(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("dataset name is required")
	}

	// Use -r to destroy recursively (including snapshots and children)
	_, err := m.exec.Output(ctx, "zfs", "destroy", "-r", name)
	return err
}

// SetProperty sets a property on a dataset.
func (m *Manager) SetProperty(ctx context.Context, name, key, value string) error {
	if name == "" || key == "" {
		return fmt.Errorf("dataset name and property key are required")
	}

	_, err := m.exec.Output(ctx, "zfs", "set", fmt.Sprintf("%s=%s", key, value), name)
	return err
}

// formatSize formats bytes to human-readable ZFS size format.
func formatSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%dT", bytes/TB)
	case bytes >= GB:
		return fmt.Sprintf("%dG", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%dM", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%dK", bytes/KB)
	default:
		return fmt.Sprintf("%d", bytes)
	}
}
