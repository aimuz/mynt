package monitor

import (
	"context"
	"fmt"

	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/zfs"
)

// ZFSScanner monitors ZFS pool health.
type ZFSScanner struct {
	bus *event.Bus
	mgr *zfs.Manager
}

// NewZFSScanner creates a ZFS scanner that publishes to the event bus.
func NewZFSScanner(bus *event.Bus, mgr *zfs.Manager) *ZFSScanner {
	return &ZFSScanner{
		bus: bus,
		mgr: mgr,
	}
}

// Scan checks ZFS pool health and publishes events.
func (s *ZFSScanner) Scan(ctx context.Context) error {
	pools, err := s.mgr.ListPools(ctx)
	if err != nil {
		return fmt.Errorf("zfs scan failed: %w", err)
	}

	for _, pool := range pools {
		if pool.Health != zfs.PoolOnline {
			s.bus.Publish(event.Event{
				Type: event.PoolDegraded,
				Data: pool,
			})
		}
	}

	return nil
}
