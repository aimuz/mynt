package scheduler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// runRetentionCleanup checks all policies and removes expired snapshots.
func (s *Scheduler) runRetentionCleanup(ctx context.Context) {
	s.logger.Debug("running retention cleanup")

	policies, err := s.policyRepo.List()
	if err != nil {
		s.logger.Error("failed to list policies for retention cleanup", "error", err)
		return
	}

	for _, policy := range policies {
		if !policy.Enabled || policy.Retention == "forever" {
			continue
		}

		retention, err := parseRetention(policy.Retention)
		if err != nil {
			s.logger.Error("invalid retention format",
				"policy", policy.Name,
				"retention", policy.Retention,
				"error", err)
			continue
		}

		s.cleanupPolicySnapshots(ctx, policy.Name, policy.Datasets, retention)
	}
}

// cleanupPolicySnapshots removes snapshots older than the retention period.
func (s *Scheduler) cleanupPolicySnapshots(ctx context.Context, policyName string, datasets []string, retention time.Duration) {
	cutoff := time.Now().Add(-retention)
	prefix := fmt.Sprintf("auto-%s-", policyName)

	for _, dataset := range datasets {
		snapshots, err := s.zfsMgr.ListSnapshots(ctx, dataset)
		if err != nil {
			s.logger.Error("failed to list snapshots for cleanup",
				"dataset", dataset,
				"error", err)
			continue
		}

		for _, snap := range snapshots {
			// Only clean up snapshots created by this policy
			parts := strings.Split(snap.Name, "@")
			if len(parts) != 2 {
				continue
			}
			snapName := parts[1]

			if !strings.HasPrefix(snapName, prefix) {
				continue
			}

			// Parse timestamp from snapshot name
			// Format: auto-{policyName}-{YYYYMMDD-HHMMSS}
			timestampStr := strings.TrimPrefix(snapName, prefix)
			snapTime, err := parseSnapshotTimestamp(timestampStr)
			if err != nil {
				s.logger.Debug("could not parse snapshot timestamp",
					"snapshot", snap.Name,
					"error", err)
				continue
			}

			if snapTime.Before(cutoff) {
				s.logger.Info("deleting expired snapshot",
					"snapshot", snap.Name,
					"policy", policyName,
					"age", time.Since(snapTime).Round(time.Hour))

				if err := s.zfsMgr.DestroySnapshot(ctx, snap.Name); err != nil {
					s.logger.Error("failed to delete expired snapshot",
						"snapshot", snap.Name,
						"error", err)
				}
			}
		}
	}
}

// parseRetention parses retention strings like "24h", "7d", "30d", "365d".
func parseRetention(retention string) (time.Duration, error) {
	retention = strings.TrimSpace(strings.ToLower(retention))

	if retention == "forever" {
		return 0, fmt.Errorf("forever retention should not be parsed")
	}

	// Match number followed by unit
	re := regexp.MustCompile(`^(\d+)([hd])$`)
	matches := re.FindStringSubmatch(retention)
	if matches == nil {
		return 0, fmt.Errorf("invalid retention format: %s", retention)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown retention unit: %s", unit)
	}
}

// parseSnapshotTimestamp parses timestamp from snapshot name format YYYYMMDD-HHMMSS.
func parseSnapshotTimestamp(timestamp string) (time.Time, error) {
	return time.Parse("20060102-150405", timestamp)
}
