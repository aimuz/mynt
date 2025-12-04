package disk

import (
	"context"
	"encoding/json"
	"fmt"
)

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

// list returns all physical disks on Linux.
// Optimized to batch all detections in a single system call.
func (m *Manager) list(ctx context.Context) ([]Info, error) {
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
