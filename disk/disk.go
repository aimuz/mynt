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
	Name   string `json:"name"`   // e.g., "sda", "nvme0n1"
	Path   string `json:"path"`   // e.g., "/dev/sda"
	Model  string `json:"model"`  // e.g., "Samsung SSD 860"
	Serial string `json:"serial"` // Unique serial number
	Size   uint64 `json:"size"`   // Size in bytes
	Type   Type   `json:"type"`   // Disk technology
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

// listLinux uses lsblk to find disks on Linux.
func (m *Manager) listLinux(ctx context.Context) ([]Info, error) {
	out, err := m.exec.Output(ctx, "lsblk", "-J", "-b", "-o", "NAME,PATH,MODEL,SERIAL,SIZE,ROTA,TYPE")
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	var result struct {
		BlockDevices []struct {
			Name   string `json:"name"`
			Path   string `json:"path"`
			Model  string `json:"model"`
			Serial string `json:"serial"`
			Size   uint64 `json:"size"`
			Rota   bool   `json:"rota"` // true = HDD, false = SSD
			Type   string `json:"type"` // disk, part, rom
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse lsblk output: %w", err)
	}

	var disks []Info
	for _, bd := range result.BlockDevices {
		if bd.Type != "disk" {
			continue
		}

		diskType := SSD
		if bd.Rota {
			diskType = HDD
		}

		disks = append(disks, Info{
			Name:   bd.Name,
			Path:   bd.Path,
			Model:  bd.Model,
			Serial: bd.Serial,
			Size:   bd.Size,
			Type:   diskType,
		})
	}

	return disks, nil
}

// listMac returns mock data for development on macOS.
func (m *Manager) listMac(ctx context.Context) ([]Info, error) {
	// For development, return mock data
	// In production, this would use diskutil
	return []Info{
		{
			Name:   "disk0",
			Path:   "/dev/disk0",
			Model:  "APPLE SSD AP0512M",
			Serial: "C02X...",
			Size:   500107862016,
			Type:   SSD,
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
			Name:   "disk3",
			Path:   "/dev/disk3",
			Model:  "Samsung 970 EVO",
			Serial: "S4X...",
			Size:   1000204886016,
			Type:   NVMe,
		},
	}, nil
}
