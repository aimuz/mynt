package disk

import (
	"context"
	"testing"
)

func TestEnrichFromCache(t *testing.T) {
	tests := []struct {
		name        string
		cache       *CachedSmart
		wantHealth  SmartHealth
		wantStatus  Status
		wantTemp    int
	}{
		{
			name: "healthy",
			cache: &CachedSmart{
				Passed:              true,
				Temperature:         35,
				ReallocatedSectors:  0,
				PendingSectors:      0,
				UncorrectableErrors: 0,
			},
			wantHealth: SmartHealthGood,
			wantStatus: StatusHealthy,
			wantTemp:   35,
		},
		{
			name: "warning_reallocated",
			cache: &CachedSmart{
				Passed:              true,
				Temperature:         42,
				ReallocatedSectors:  5,
				PendingSectors:      0,
				UncorrectableErrors: 0,
			},
			wantHealth: SmartHealthWarning,
			wantStatus: StatusWarning,
			wantTemp:   42,
		},
		{
			name: "warning_pending",
			cache: &CachedSmart{
				Passed:              true,
				Temperature:         38,
				ReallocatedSectors:  0,
				PendingSectors:      3,
				UncorrectableErrors: 0,
			},
			wantHealth: SmartHealthWarning,
			wantStatus: StatusWarning,
			wantTemp:   38,
		},
		{
			name: "failed",
			cache: &CachedSmart{
				Passed:              false,
				Temperature:         55,
				ReallocatedSectors:  100,
				PendingSectors:      50,
				UncorrectableErrors: 10,
			},
			wantHealth: SmartHealthFailed,
			wantStatus: StatusFailed,
			wantTemp:   55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &Info{}
			enrichFromCache(info, tt.cache)

			if info.SmartHealth != tt.wantHealth {
				t.Errorf("SmartHealth = %v, want %v", info.SmartHealth, tt.wantHealth)
			}
			if info.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", info.Status, tt.wantStatus)
			}
			if info.Temperature != tt.wantTemp {
				t.Errorf("Temperature = %v, want %v", info.Temperature, tt.wantTemp)
			}
		})
	}
}

func TestManagerOptions(t *testing.T) {
	tests := []struct {
		name                   string
		opts                   []ManagerOption
		wantIncludeLoopDevices bool
		wantCache              bool
	}{
		{
			name:                   "default",
			opts:                   nil,
			wantIncludeLoopDevices: false,
			wantCache:              false,
		},
		{
			name:                   "with_loop_devices",
			opts:                   []ManagerOption{WithLoopDevices()},
			wantIncludeLoopDevices: true,
			wantCache:              false,
		},
		{
			name:                   "with_cache",
			opts:                   []ManagerOption{WithSmartCache(&mockSmartCache{})},
			wantIncludeLoopDevices: false,
			wantCache:              true,
		},
		{
			name:                   "with_both",
			opts:                   []ManagerOption{WithLoopDevices(), WithSmartCache(&mockSmartCache{})},
			wantIncludeLoopDevices: true,
			wantCache:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.opts...)
			if m.includeLoopDevices != tt.wantIncludeLoopDevices {
				t.Errorf("includeLoopDevices = %v, want %v", m.includeLoopDevices, tt.wantIncludeLoopDevices)
			}
			if (m.cache != nil) != tt.wantCache {
				t.Errorf("cache set = %v, want %v", m.cache != nil, tt.wantCache)
			}
		})
	}
}

// mockSmartCache is a simple mock for testing.
type mockSmartCache struct {
	data map[string]*CachedSmart
}

func (m *mockSmartCache) GetSmart(name string) (*CachedSmart, error) {
	if m.data == nil {
		return nil, nil
	}
	return m.data[name], nil
}

func (m *mockSmartCache) ListSmart() (map[string]*CachedSmart, error) {
	if m.data == nil {
		return make(map[string]*CachedSmart), nil
	}
	return m.data, nil
}

func TestManagerList_WithCache(t *testing.T) {
	cache := &mockSmartCache{
		data: map[string]*CachedSmart{
			"sda": {
				Passed:              true,
				Temperature:         35,
				ReallocatedSectors:  0,
				PendingSectors:      0,
				UncorrectableErrors: 0,
			},
		},
	}

	m := NewManager(WithSmartCache(cache))

	// Test that cache is properly configured
	if m.cache == nil {
		t.Fatal("cache should be set")
	}

	// We can't test actual disk listing without real hardware,
	// but we can verify the manager is properly configured
	ctx := context.Background()
	_, err := m.List(ctx)
	// Error is expected since we don't have real disks
	// We're just testing the code path doesn't panic
	_ = err
}

func TestType_Constants(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{HDD, "HDD"},
		{SSD, "SSD"},
		{NVMe, "NVMe"},
		{USB, "USB"},
		{Unknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("Type = %v, want %v", tt.typ, tt.want)
			}
		})
	}
}

func TestUsageType_Constants(t *testing.T) {
	tests := []struct {
		usage UsageType
		want  string
	}{
		{UsageTypeNone, ""},
		{UsageTypeZFSMember, "zfs_member"},
		{UsageTypeFormatted, "formatted"},
		{UsageTypePartitions, "has_partitions"},
		{UsageTypeSystem, "system_disk"},
	}

	for _, tt := range tests {
		t.Run(string(tt.usage), func(t *testing.T) {
			if string(tt.usage) != tt.want {
				t.Errorf("UsageType = %v, want %v", tt.usage, tt.want)
			}
		})
	}
}

func TestStatus_Constants(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusHealthy, "healthy"},
		{StatusWarning, "warning"},
		{StatusFailed, "failed"},
		{StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

func TestSmartHealth_Constants(t *testing.T) {
	tests := []struct {
		health SmartHealth
		want   string
	}{
		{SmartHealthGood, "good"},
		{SmartHealthWarning, "warning"},
		{SmartHealthFailed, "failed"},
		{SmartHealthUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.health), func(t *testing.T) {
			if string(tt.health) != tt.want {
				t.Errorf("SmartHealth = %v, want %v", tt.health, tt.want)
			}
		})
	}
}
