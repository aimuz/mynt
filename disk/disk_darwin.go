package disk

import (
	"context"
)

// list returns mock disk data for development on macOS.
// In production, this would use diskutil to query actual disk information.
func (m *Manager) list(ctx context.Context) ([]Info, error) {
	return []Info{
		{
			Name:   "disk0",
			Path:   "/dev/disk0",
			Model:  "APPLE SSD AP0512M",
			Serial: "C02X...",
			Size:   500107862016,
			Type:   SSD,
			InUse:  true,
			Usage: &UsageInfo{
				Type:   UsageTypeSystem,
				Params: map[string]string{"fstype": "APFS"},
			},
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
			InUse:  true,
			Usage: &UsageInfo{
				Type:   UsageTypeZFSMember,
				Params: map[string]string{"pool": "tank"},
			},
		},
	}, nil
}
