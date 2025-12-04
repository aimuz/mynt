// Package disk provides disk discovery and monitoring.
// It replaces internal/hardware with a simpler, more focused API.
package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

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
	if runtime.GOOS == "darwin" {
		return m.listMac(ctx)
	}
	return m.listLinux(ctx)
}

// lsblkDevice represents a block device from lsblk output.
type lsblkDevice struct {
	Name     string        `json:"name"`
	Path     string        `json:"path"`
	Model    string        `json:"model"`
	Serial   string        `json:"serial"`
	Size     uint64        `json:"size"`
	Rota     bool          `json:"rota"` // true = HDD, false = SSD
	Type     string        `json:"type"` // disk, part, rom
	Fstype   string        `json:"fstype"`
	Label    string        `json:"label"`
	Children []lsblkDevice `json:"children,omitempty"`
}

// listLinux uses lsblk to find disks on Linux.
// Optimized to batch all detections in a single system call.
func (m *Manager) listLinux(ctx context.Context) ([]Info, error) {
	out, err := m.exec.Output(ctx, "lsblk", "-J", "-b", "-o", "NAME,PATH,MODEL,SERIAL,SIZE,ROTA,TYPE,FSTYPE,LABEL")
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	var result struct {
		BlockDevices []lsblkDevice `json:"blockdevices"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse lsblk output: %w", err)
	}

	disks := make([]Info, 0, len(result.BlockDevices))
	for _, bd := range result.BlockDevices {
		if bd.Type != "disk" {
			continue
		}

		info := Info{
			Name:   bd.Name,
			Path:   bd.Path,
			Model:  bd.Model,
			Serial: bd.Serial,
			Size:   bd.Size,
			Type:   detectDiskType(bd.Name, bd.Rota),
		}

		// Detect usage status from lsblk data
		detectUsageStatus(&info, &bd)

		disks = append(disks, info)
	}

	return disks, nil
}

// detectDiskType infers disk technology from device name and rotation flag.
func detectDiskType(name string, rota bool) Type {
	// NVMe devices have names like nvme0n1, nvme1n1, etc.
	if len(name) >= 4 && name[:4] == "nvme" {
		return NVMe
	}
	if rota {
		return HDD
	}
	return SSD
}

// detectUsageStatus determines if a disk is in use and why.
func detectUsageStatus(info *Info, bd *lsblkDevice) {
	if bd.Fstype != "" {
		info.InUse = true
		if bd.Fstype == "zfs_member" {
			if bd.Label != "" {
				info.UsageReason = fmt.Sprintf("ZFS Pool Member (%s)", bd.Label)
			} else {
				info.UsageReason = "ZFS Pool Member"
			}
		} else {
			info.UsageReason = fmt.Sprintf("Formatted (%s)", bd.Fstype)
		}
		return
	}

	// Check for partition children
	for _, child := range bd.Children {
		if child.Type == "part" {
			info.InUse = true
			info.UsageReason = "Has Partitions"
			return
		}
	}
}

// listMac returns mock data for development on macOS.
func (m *Manager) listMac(ctx context.Context) ([]Info, error) {
	// For development, return mock data
	// In production, this would use diskutil
	return []Info{
		{
			Name:        "disk0",
			Path:        "/dev/disk0",
			Model:       "APPLE SSD AP0512M",
			Serial:      "C02X...",
			Size:        500107862016,
			Type:        SSD,
			InUse:       true,
			UsageReason: "System Disk (APFS)",
		},
		{
			Name:   "disk2",
			Path:   "/dev/disk2",
			Model:  "WD Red Plus",
			Serial: "WD-WCC...",
			Size:   4000787030016,
			Type:   HDD,
		},
		{
			Name:        "disk3",
			Path:        "/dev/disk3",
			Model:       "Samsung 970 EVO",
			Serial:      "S4X...",
			Size:        1000204886016,
			Type:        NVMe,
			InUse:       true,
			UsageReason: "ZFS Pool Member (tank)",
		},
	}, nil
}
