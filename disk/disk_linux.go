package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// lsblkDevice represents a block device from lsblk output.
type lsblkDevice struct {
	Name     string        `json:"name"`
	Path     string        `json:"path"`
	Model    string        `json:"model"`
	Serial   string        `json:"serial"`
	Size     uint64        `json:"size"`
	Rota     bool          `json:"rota"`
	Type     string        `json:"type"`
	Fstype   string        `json:"fstype"`
	Label    string        `json:"label"`
	Children []lsblkDevice `json:"children,omitempty"`
}

// listBasic returns all physical disks without SMART data (fast).
func (m *Manager) listBasic(ctx context.Context) ([]Info, error) {
	out, err := m.exec.Output(ctx, "lsblk", "-J", "-b", "-o", "NAME,PATH,MODEL,SERIAL,SIZE,ROTA,TYPE,FSTYPE,LABEL")
	if err != nil {
		return nil, fmt.Errorf("lsblk: %w", err)
	}

	var result struct {
		BlockDevices []lsblkDevice `json:"blockdevices"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse lsblk: %w", err)
	}

	var disks []Info
	for _, d := range result.BlockDevices {
		if d.Type != "disk" && !(m.includeLoopDevices && d.Type == "loop") {
			continue
		}

		// Exclude ZFS volumes (e.g. zd0, zd16) which appear as disks
		if strings.HasPrefix(d.Name, "zd") {
			continue
		}

		info := Info{
			Name:        d.Name,
			Path:        d.Path,
			Model:       d.Model,
			Serial:      d.Serial,
			Size:        d.Size,
			Type:        diskType(d.Name, d.Rota),
			Status:      StatusUnknown,
			SmartHealth: SmartHealthUnknown,
		}

		setUsage(&info, &d)
		disks = append(disks, info)
	}
	return disks, nil
}

// diskType infers disk technology from device name and rotation flag.
func diskType(name string, rota bool) Type {
	if strings.HasPrefix(name, "nvme") {
		return NVMe
	}
	if rota {
		return HDD
	}
	return SSD
}

// setUsage determines if a disk is in use and why.
func setUsage(info *Info, d *lsblkDevice) {
	if d.Fstype != "" {
		info.InUse = true
		if d.Fstype == "zfs_member" {
			info.Usage = &UsageInfo{Type: UsageTypeZFSMember}
			if d.Label != "" {
				info.Usage.Params = map[string]string{"pool": d.Label}
				info.Pool = d.Label
			}
		} else {
			info.Usage = &UsageInfo{
				Type:   UsageTypeFormatted,
				Params: map[string]string{"fstype": d.Fstype},
			}
		}
		return
	}

	for _, c := range d.Children {
		if c.Type == "part" {
			info.InUse = true
			info.Usage = &UsageInfo{Type: UsageTypePartitions}
			return
		}
	}
}
