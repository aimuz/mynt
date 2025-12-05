// Package scheduler provides cron-based snapshot policy execution.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/zfs"
)

// Scheduler manages automatic snapshot creation based on policies.
type Scheduler struct {
	cron       *cron.Cron
	policyRepo *store.SnapshotPolicyRepo
	zfsMgr     *zfs.Manager
	logger     *slog.Logger

	mu       sync.RWMutex
	entryIDs map[int64]cron.EntryID // policyID -> cronEntryID
}

// New creates a new Scheduler.
func New(policyRepo *store.SnapshotPolicyRepo, zfsMgr *zfs.Manager) *Scheduler {
	return &Scheduler{
		cron:       cron.New(cron.WithSeconds()),
		policyRepo: policyRepo,
		zfsMgr:     zfsMgr,
		logger:     slog.Default(),
		entryIDs:   make(map[int64]cron.EntryID),
	}
}

// Start begins the scheduler and loads all policies.
func (s *Scheduler) Start(ctx context.Context) error {
	s.logger.Info("starting snapshot policy scheduler")

	// Load and schedule all enabled policies
	if err := s.Reload(); err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// Add retention cleanup job (runs every hour)
	_, err := s.cron.AddFunc("0 0 * * * *", func() {
		s.runRetentionCleanup(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add retention cleanup job: %w", err)
	}

	s.cron.Start()
	s.logger.Info("snapshot policy scheduler started", "policies", len(s.entryIDs))

	return nil
}

// Stop halts the scheduler gracefully.
func (s *Scheduler) Stop() {
	s.logger.Info("stopping snapshot policy scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("snapshot policy scheduler stopped")
}

// Reload reloads all policies from the database.
// Call this after creating, updating, or deleting a policy.
func (s *Scheduler) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove all existing policy jobs
	for policyID, entryID := range s.entryIDs {
		s.cron.Remove(entryID)
		delete(s.entryIDs, policyID)
	}

	// Load policies from database
	policies, err := s.policyRepo.List()
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	// Schedule enabled policies
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		if err := s.schedulePolicy(policy); err != nil {
			s.logger.Error("failed to schedule policy",
				"policy", policy.Name,
				"schedule", policy.Schedule,
				"error", err)
			continue
		}
	}

	s.logger.Info("policies reloaded", "scheduled", len(s.entryIDs))
	return nil
}

// schedulePolicy adds a policy to the cron scheduler.
func (s *Scheduler) schedulePolicy(policy store.SnapshotPolicy) error {
	// Convert schedule to cron format
	schedule := convertSchedule(policy.Schedule)

	entryID, err := s.cron.AddFunc(schedule, func() {
		s.executePolicy(policy)
	})
	if err != nil {
		return fmt.Errorf("invalid schedule %q: %w", policy.Schedule, err)
	}

	s.entryIDs[policy.ID] = entryID
	s.logger.Debug("scheduled policy",
		"policy", policy.Name,
		"schedule", schedule,
		"datasets", len(policy.Datasets))

	return nil
}

// convertSchedule converts user-friendly schedules to cron format.
// robfig/cron uses 6 fields: second minute hour day month weekday
func convertSchedule(schedule string) string {
	switch schedule {
	case "@hourly":
		return "0 0 * * * *" // At minute 0 of every hour
	case "@daily":
		return "0 0 0 * * *" // At midnight
	case "@weekly":
		return "0 0 0 * * 0" // Sunday at midnight
	case "@monthly":
		return "0 0 0 1 * *" // 1st of each month at midnight
	default:
		// Assume it's already a valid cron expression
		// If it's a 5-field expression, prepend "0" for seconds
		if len(schedule) > 0 && schedule[0] != '@' {
			fields := 0
			for _, c := range schedule {
				if c == ' ' {
					fields++
				}
			}
			if fields == 4 {
				// 5-field cron (minute hour day month weekday)
				return "0 " + schedule
			}
		}
		return schedule
	}
}

// executePolicy creates snapshots for all datasets in a policy.
func (s *Scheduler) executePolicy(policy store.SnapshotPolicy) {
	ctx := context.Background()
	timestamp := time.Now().Format("20060102-150405")
	snapshotName := fmt.Sprintf("auto-%s-%s", policy.Name, timestamp)

	s.logger.Info("executing snapshot policy",
		"policy", policy.Name,
		"datasets", len(policy.Datasets))

	for _, dataset := range policy.Datasets {
		req := zfs.CreateSnapshotRequest{
			Dataset: dataset,
			Name:    snapshotName,
		}

		snapshot, err := s.zfsMgr.CreateSnapshot(ctx, req)
		if err != nil {
			s.logger.Error("failed to create snapshot",
				"policy", policy.Name,
				"dataset", dataset,
				"error", err)
			continue
		}

		// Update source to indicate this was created by a policy
		snapshot.Source = fmt.Sprintf("policy:%s", policy.Name)

		s.logger.Info("snapshot created by policy",
			"policy", policy.Name,
			"snapshot", snapshot.Name)
	}
}
