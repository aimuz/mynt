// Package disk provides disk discovery and monitoring.
package disk

import (
	"context"

	"go.aimuz.me/mynt/logger"
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
type UsageInfo struct {
	Type   UsageType         `json:"type"`
	Params map[string]string `json:"params,omitempty"`
}

// Status represents the health status of a disk.
type Status string

const (
	StatusHealthy Status = "healthy"
	StatusWarning Status = "warning"
	StatusFailed  Status = "failed"
	StatusUnknown Status = "unknown"
)

// SmartHealth represents the S.M.A.R.T. health status.
type SmartHealth string

const (
	SmartHealthGood    SmartHealth = "good"
	SmartHealthWarning SmartHealth = "warning"
	SmartHealthFailed  SmartHealth = "failed"
	SmartHealthUnknown SmartHealth = "unknown"
)

// Info represents a physical disk.
type Info struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Model       string      `json:"model"`
	Serial      string      `json:"serial"`
	Size        uint64      `json:"size"`
	Type        Type        `json:"type"`
	InUse       bool        `json:"in_use"`
	Usage       *UsageInfo  `json:"usage,omitempty"`
	Slot        string      `json:"slot,omitempty"`
	Pool        string      `json:"pool,omitempty"`
	Status      Status      `json:"status"`
	SmartHealth SmartHealth `json:"smart_health"`
	Temperature int         `json:"temperature"`
}

// SmartCache provides cached SMART data.
type SmartCache interface {
	GetSmart(name string) (*CachedSmart, error)
	ListSmart() (map[string]*CachedSmart, error)
}

// CachedSmart holds cached SMART data.
type CachedSmart struct {
	Passed              bool
	Temperature         int
	ReallocatedSectors  int64
	PendingSectors      int64
	UncorrectableErrors int64
}

// Manager handles disk operations.
type Manager struct {
	exec               sysexec.Executor
	includeLoopDevices bool
	cache              SmartCache
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithLoopDevices enables loop device detection.
func WithLoopDevices() ManagerOption {
	return func(m *Manager) { m.includeLoopDevices = true }
}

// WithSmartCache sets the SMART cache source.
func WithSmartCache(c SmartCache) ManagerOption {
	return func(m *Manager) { m.cache = c }
}

// NewManager creates a new disk manager.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{exec: sysexec.NewExecutor()}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// List returns all physical disks with cached SMART data.
func (m *Manager) List(ctx context.Context) ([]Info, error) {
	disks, err := m.listBasic(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich with cached SMART data if available
	if m.cache != nil {
		smartMap, err := m.cache.ListSmart()
		if err != nil {
			logger.Debug("failed to load SMART cache", "error", err)
		}
		for i := range disks {
			if s, ok := smartMap[disks[i].Name]; ok {
				enrichFromCache(&disks[i], s)
			}
		}
	}

	return disks, nil
}

// ListBasic returns disks without SMART data (fast).
func (m *Manager) ListBasic(ctx context.Context) ([]Info, error) {
	return m.listBasic(ctx)
}

// enrichFromCache populates Info from cached SMART data.
func enrichFromCache(info *Info, s *CachedSmart) {
	info.Temperature = s.Temperature

	if s.Passed {
		info.SmartHealth = SmartHealthGood
		if s.ReallocatedSectors > 0 || s.PendingSectors > 0 {
			info.SmartHealth = SmartHealthWarning
		}
	} else {
		info.SmartHealth = SmartHealthFailed
	}

	switch info.SmartHealth {
	case SmartHealthGood:
		info.Status = StatusHealthy
	case SmartHealthWarning:
		info.Status = StatusWarning
	case SmartHealthFailed:
		info.Status = StatusFailed
	}
}
