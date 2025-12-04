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
	UseCase    UseCaseTemplate   `json:"use_case"`   // template to apply
	QuotaMode  string            `json:"quota_mode"` // "fixed", "flexible"
	Quota      uint64            `json:"quota,omitempty"`
	Properties map[string]string `json:"properties"` // optional ZFS properties (overrides template)
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

	// Apply use-case template properties
	properties := GetTemplateProperties(req.UseCase)

	// Merge user-provided properties (overrides template)
	for k, v := range req.Properties {
		properties[k] = v
	}

	// Apply quota if specified
	if req.Quota > 0 {
		if req.QuotaMode == "fixed" {
			properties["reservation"] = fmt.Sprintf("%d", req.Quota)
			properties["quota"] = fmt.Sprintf("%d", req.Quota)
		} else {
			// Flexible mode: only set quota, no reservation
			properties["quota"] = fmt.Sprintf("%d", req.Quota)
		}
	}

	var err error

	if req.Type == "volume" {
		if req.Size == 0 {
			return fmt.Errorf("size is required for volumes")
		}
		_, err = gozfs.CreateVolume(req.Name, req.Size, properties)
	} else {
		_, err = gozfs.CreateFilesystem(req.Name, properties)
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
	// Get quota and reservation
	if quota, err := gozfsDataset.GetProperty("quota"); err == nil && quota != "0" && quota != "none" {
		var q uint64
		fmt.Sscanf(quota, "%d", &q)
		dataset.Quota = q
	}
	if res, err := gozfsDataset.GetProperty("reservation"); err == nil && res != "0" && res != "none" {
		var r uint64
		fmt.Sscanf(res, "%d", &r)
		dataset.Reservation = r
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

// SetQuota sets a quota on a dataset.
func (m *Manager) SetQuota(ctx context.Context, name string, quota uint64) error {
	return m.SetProperty(ctx, name, "quota", fmt.Sprintf("%d", quota))
}

// SetReservation sets a reservation on a dataset.
func (m *Manager) SetReservation(ctx context.Context, name string, reservation uint64) error {
	return m.SetProperty(ctx, name, "reservation", fmt.Sprintf("%d", reservation))
}

// GetTemplateProperties returns ZFS properties for a given use-case template.
func GetTemplateProperties(useCase UseCaseTemplate) map[string]string {
	switch useCase {
	case UseCaseMedia:
		return map[string]string{
			"recordsize":  "1M",  // Large block size for large files
			"compression": "lz4", // Fast compression
			"atime":       "off", // No access time updates
		}
	case UseCaseSurveillance:
		return map[string]string{
			"recordsize":  "1M", // Large block size for video files
			"compression": "lz4",
			"atime":       "off",
			"sync":        "standard", // Balance between performance and safety
		}
	case UseCaseVM:
		return map[string]string{
			"recordsize":  "64K", // Smaller block size for random I/O
			"compression": "lz4",
			"sync":        "disabled", // Maximum performance (use with caution)
		}
	case UseCaseDatabase:
		return map[string]string{
			"recordsize":  "16K", // Match typical database page size
			"compression": "lz4",
			"logbias":     "latency", // Optimize for latency
			"sync":        "always",  // Data integrity priority
		}
	case UseCaseGeneral:
		fallthrough
	default:
		return map[string]string{
			"recordsize":  "128K", // Default ZFS recordsize
			"compression": "lz4",
		}
	}
}
