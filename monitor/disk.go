package monitor

import (
	"context"
	"fmt"
	"time"

	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/store"
)

// DiskScanner monitors disk changes (fast, runs frequently).
type DiskScanner struct {
	bus     *event.Bus
	repo    *store.DiskRepo
	diskMgr *disk.Manager
}

// NewDiskScanner creates a disk scanner that publishes to the event bus.
func NewDiskScanner(bus *event.Bus, repo *store.DiskRepo, diskMgr *disk.Manager) *DiskScanner {
	return &DiskScanner{
		bus:     bus,
		repo:    repo,
		diskMgr: diskMgr,
	}
}

// Scan checks for disk changes (does NOT collect SMART data).
func (s *DiskScanner) Scan(ctx context.Context) error {
	current, err := s.diskMgr.ListBasic(ctx)
	if err != nil {
		return fmt.Errorf("disk scan: %w", err)
	}

	known, err := s.repo.ListAttached()
	if err != nil {
		return fmt.Errorf("list known disks: %w", err)
	}

	currentMap := make(map[string]disk.Info)
	knownMap := make(map[string]store.DiskState)

	for _, d := range current {
		currentMap[d.Serial] = d
	}
	for _, d := range known {
		knownMap[d.Serial] = d
	}

	for serial, d := range currentMap {
		if _, exists := knownMap[serial]; !exists {
			s.bus.Publish(event.Event{Type: event.DiskAdded, Data: d})
		}
		if err := s.repo.Save(d); err != nil {
			fmt.Printf("Warning: failed to save disk %s: %v\n", d.Name, err)
		}
	}

	for serial, d := range knownMap {
		if _, exists := currentMap[serial]; !exists {
			s.bus.Publish(event.Event{Type: event.DiskRemoved, Data: d.ToInfo()})
			if err := s.repo.MarkDetached(d.Name, d.Serial); err != nil {
				fmt.Printf("Warning: failed to mark disk %s as detached: %v\n", d.Name, err)
			}
			s.repo.DeleteSmart(d.Name)
		}
	}

	return nil
}

// SmartScanner collects SMART data (slow, runs less frequently).
type SmartScanner struct {
	bus        *event.Bus
	repo       *store.DiskRepo
	diskMgr    *disk.Manager
	lastUpdate time.Time
	interval   time.Duration
}

// NewSmartScanner creates a SMART data collector.
// interval specifies how often to actually collect SMART data.
func NewSmartScanner(bus *event.Bus, repo *store.DiskRepo, diskMgr *disk.Manager, interval time.Duration) *SmartScanner {
	return &SmartScanner{
		bus:      bus,
		repo:     repo,
		diskMgr:  diskMgr,
		interval: interval,
	}
}

// Scan collects SMART data for all attached disks.
func (s *SmartScanner) Scan(ctx context.Context) error {
	// Check if enough time has passed since last update
	if time.Since(s.lastUpdate) < s.interval {
		return nil
	}
	s.lastUpdate = time.Now()

	disks, err := s.diskMgr.ListBasic(ctx)
	if err != nil {
		return fmt.Errorf("smart scan: %w", err)
	}

	for _, d := range disks {
		s.collectSmart(ctx, d.Name)
	}

	return nil
}

func (s *SmartScanner) collectSmart(ctx context.Context, name string) {
	report, err := s.diskMgr.SmartDetails(ctx, name)
	if err != nil {
		return
	}

	if err := s.repo.SaveSmart(report); err != nil {
		fmt.Printf("Warning: failed to cache SMART for %s: %v\n", name, err)
		return
	}

	if !report.Passed {
		s.bus.Publish(event.Event{
			Type: event.SmartFailed,
			Data: map[string]any{"disk": name, "report": report},
		})
	}
}
