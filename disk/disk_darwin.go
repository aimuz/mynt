package disk

import (
	"context"
)

// listBasic returns mock disk data for development on macOS.
func (m *Manager) listBasic(ctx context.Context) ([]Info, error) {
	return []Info{
		{
			Name:        "disk0",
			Path:        "/dev/disk0",
			Model:       "APPLE SSD AP0512M",
			Serial:      "C02X1234567",
			Size:        500107862016,
			Type:        SSD,
			InUse:       true,
			Usage:       &UsageInfo{Type: UsageTypeSystem, Params: map[string]string{"fstype": "APFS"}},
			Slot:        "Bay 1",
			Status:      StatusHealthy,
			SmartHealth: SmartHealthGood,
			Temperature: 38,
		},
		{
			Name:        "disk2",
			Path:        "/dev/disk2",
			Model:       "WD Red Plus 4TB",
			Serial:      "WD-WCC4N1234567",
			Size:        4000787030016,
			Type:        HDD,
			Slot:        "Bay 2",
			Status:      StatusHealthy,
			SmartHealth: SmartHealthGood,
			Temperature: 35,
		},
		{
			Name:        "disk3",
			Path:        "/dev/disk3",
			Model:       "Samsung 970 EVO Plus",
			Serial:      "S4XXNF0M123456",
			Size:        1000204886016,
			Type:        NVMe,
			InUse:       true,
			Usage:       &UsageInfo{Type: UsageTypeZFSMember, Params: map[string]string{"pool": "tank"}},
			Slot:        "NVMe Slot 1",
			Pool:        "tank",
			Status:      StatusHealthy,
			SmartHealth: SmartHealthGood,
			Temperature: 42,
		},
		{
			Name:        "disk4",
			Path:        "/dev/disk4",
			Model:       "WD Blue 2TB",
			Serial:      "WD-WCC1234567890",
			Size:        2000398934016,
			Type:        HDD,
			Slot:        "Bay 3",
			Status:      StatusWarning,
			SmartHealth: SmartHealthWarning,
			Temperature: 48,
		},
	}, nil
}
