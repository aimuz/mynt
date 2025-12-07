// Package monitor provides unified system monitoring.
// It replaces the previous dual-system of SystemCoordinator.startMonitoring and sys.Worker.
package monitor

import (
	"context"
	"sync"
	"time"

	"go.aimuz.me/mynt/logger"
)

// Scanner represents a component that can scan for changes.
type Scanner interface {
	Scan(ctx context.Context) error
}

// Monitor coordinates all system scanners.
type Monitor struct {
	scanners []Scanner
	interval time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New creates a new monitor with the given scanners and interval.
func New(scanners []Scanner, interval time.Duration) *Monitor {
	return &Monitor{
		scanners: scanners,
		interval: interval,
	}
}

// Start begins monitoring. It runs until Stop is called.
func (m *Monitor) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)

	logger.Info("monitoring started", "scanners", len(m.scanners), "interval", m.interval)

	m.wg.Go(func() {
		m.run(ctx)
	})
}

// Stop halts monitoring and waits for completion.
func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
}

func (m *Monitor) run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Run immediately on start
	m.scan(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.scan(ctx)
		}
	}
}

func (m *Monitor) scan(ctx context.Context) {
	for _, scanner := range m.scanners {
		if err := scanner.Scan(ctx); err != nil {
			// Log error but continue with other scanners
			// In production, use structured logging
			logger.Error("failed to scan", "error", err)
		}
	}
}
