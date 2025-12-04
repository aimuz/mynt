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

// Info represents a physical disk.
type Info struct {
	Name        string `json:"name"`         // e.g., "sda", "nvme0n1"
	Path        string `json:"path"`         // e.g., "/dev/sda"
	Model       string `json:"model"`        // e.g., "Samsung SSD 860"
	Serial      string `json:"serial"`       // Unique serial number
	Size        uint64 `json:"size"`         // Size in bytes
	Type        Type   `json:"type"`         // Disk technology
	InUse       bool   `json:"in_use"`       // Whether disk is currently in use
	UsageReason string `json:"usage_reason"` // Why disk is in use (if InUse is true)
}

// Manager handles disk operations.
type Manager struct {
	exec sysexec.Executor
}

// NewManager creates a new disk manager.
func NewManager() *Manager {
	return &Manager{exec: sysexec.NewExecutor()}
}

// List returns all physical disks on the system.
func (m *Manager) List(ctx context.Context) ([]Info, error) {
	return m.list(ctx)
}
