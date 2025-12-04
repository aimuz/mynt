// Package disk provides disk discovery and monitoring.
// It replaces internal/hardware with a simpler, more focused API.
package disk

import (
	"context"

	"go.aimuz.me/mynt/sysexec"
)

// Type represents the technology of a disk.
type Type string

const (
	HDD     Type = "HDD"
	SSD     Type = "SSD"
	NVMe    Type = "NVMe"
	USB     Type = "USB"
	Unknown Type = "Unknown"
)

// UsageType represents why a disk is in use.
type UsageType string

const (
	UsageTypeNone       UsageType = ""
	UsageTypeZFSMember  UsageType = "zfs_member"
	UsageTypeFormatted  UsageType = "formatted"
	UsageTypePartitions UsageType = "has_partitions"
	UsageTypeSystem     UsageType = "system_disk"
)

// UsageInfo contains structured information about disk usage.
// This enables i18n support in the frontend.
type UsageInfo struct {
	Type   UsageType         `json:"type"`             // Usage type identifier
	Params map[string]string `json:"params,omitempty"` // Parameters for the usage type
}

// Info represents a physical disk.
type Info struct {
	Name   string     `json:"name"`            // e.g., "sda", "nvme0n1"
	Path   string     `json:"path"`            // e.g., "/dev/sda"
	Model  string     `json:"model"`           // e.g., "Samsung SSD 860"
	Serial string     `json:"serial"`          // Unique serial number
	Size   uint64     `json:"size"`            // Size in bytes
	Type   Type       `json:"type"`            // Disk technology
	InUse  bool       `json:"in_use"`          // Whether disk is currently in use
	Usage  *UsageInfo `json:"usage,omitempty"` // Structured usage information for i18n
}

// Manager handles disk operations.
type Manager struct {
	exec               sysexec.Executor
	includeLoopDevices bool // Feature flag to include loop devices (useful for testing)
}

// ManagerOption is a function that configures a Manager.
type ManagerOption func(*Manager)

// WithLoopDevices enables loop device detection (useful for testing in VMs).
func WithLoopDevices() ManagerOption {
	return func(m *Manager) {
		m.includeLoopDevices = true
	}
}

// NewManager creates a new disk manager.
// Options can be passed to configure the manager:
//   - WithLoopDevices(): Enable loop device detection for testing
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{exec: sysexec.NewExecutor()}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// List returns all physical disks on the system.
func (m *Manager) List(ctx context.Context) ([]Info, error) {
	return m.list(ctx)
}
