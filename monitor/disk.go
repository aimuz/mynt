package monitor

import (
	"context"
	"fmt"

	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/store"
)

// DiskScanner monitors disk changes and SMART status.
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

// Scan checks for disk changes and publishes events.
func (s *DiskScanner) Scan(ctx context.Context) error {
	// Get current physical disks
	current, err := s.diskMgr.List(ctx)
	if err != nil {
		return fmt.Errorf("disk scan failed: %w", err)
	}

	// Get known disks from database
	known, err := s.repo.ListAttached()
	if err != nil {
		return fmt.Errorf("failed to list known disks: %w", err)
	}

	// Build maps for comparison
	currentMap := make(map[string]disk.Info)     // key: serial
	knownMap := make(map[string]store.DiskState) // key: serial

	for _, d := range current {
		currentMap[d.Serial] = d
	}
	for _, d := range known {
		knownMap[d.Serial] = d
	}

	// Check for new disks
	for serial, d := range currentMap {
		if _, exists := knownMap[serial]; !exists {
			// New disk detected
			s.bus.Publish(event.Event{
				Type: event.DiskAdded,
				Data: d,
			})
		}

		// Save/update disk state
		if err := s.repo.Save(d); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to save disk %s: %v\n", d.Name, err)
		}

		// Check SMART health
		if err := disk.CheckHealth(ctx, d.Name); err != nil {
			s.bus.Publish(event.Event{
				Type: event.SmartFailed,
				Data: map[string]any{
					"disk":  d,
					"error": err.Error(),
				},
			})
		}
	}

	// Check for removed disks
	for serial, d := range knownMap {
		if _, exists := currentMap[serial]; !exists {
			// Disk removed
			s.bus.Publish(event.Event{
				Type: event.DiskRemoved,
				Data: d.ToInfo(),
			})

			// Mark as detached in database
			if err := s.repo.MarkDetached(d.Name, d.Serial); err != nil {
				fmt.Printf("Warning: failed to mark disk %s as detached: %v\n", d.Name, err)
			}
		}
	}

	return nil
}
