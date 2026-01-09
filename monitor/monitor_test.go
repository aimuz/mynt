package monitor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockScanner is a simple scanner for testing.
type mockScanner struct {
	mu        sync.Mutex
	scanCount int
	err       error
}

func (m *mockScanner) Scan(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scanCount++
	return m.err
}

func (m *mockScanner) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.scanCount
}

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		scanners     []Scanner
		interval     time.Duration
		wantScanners int
	}{
		{
			name:         "no_scanners",
			scanners:     nil,
			interval:     time.Second,
			wantScanners: 0,
		},
		{
			name:         "one_scanner",
			scanners:     []Scanner{&mockScanner{}},
			interval:     time.Second,
			wantScanners: 1,
		},
		{
			name:         "multiple_scanners",
			scanners:     []Scanner{&mockScanner{}, &mockScanner{}},
			interval:     time.Second,
			wantScanners: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.scanners, tt.interval)
			if m == nil {
				t.Fatal("New() returned nil")
			}
			if len(m.scanners) != tt.wantScanners {
				t.Errorf("New() scanners = %v, want %v", len(m.scanners), tt.wantScanners)
			}
			if m.interval != tt.interval {
				t.Errorf("New() interval = %v, want %v", m.interval, tt.interval)
			}
		})
	}
}

func TestMonitor_StartStop(t *testing.T) {
	scanner := &mockScanner{}
	m := New([]Scanner{scanner}, 50*time.Millisecond)

	ctx := context.Background()
	m.Start(ctx)

	// Wait for a few scan cycles
	time.Sleep(150 * time.Millisecond)

	m.Stop()

	count := scanner.count()
	// Should have scanned at least 2 times (immediate + at least one interval)
	if count < 2 {
		t.Errorf("scan count = %v, want >= 2", count)
	}
}

func TestMonitor_ScanError(t *testing.T) {
	// Scanner that returns an error
	scanner := &mockScanner{err: errors.New("scan failed")}
	m := New([]Scanner{scanner}, 50*time.Millisecond)

	ctx := context.Background()
	m.Start(ctx)

	// Wait for a few scan cycles
	time.Sleep(150 * time.Millisecond)

	m.Stop()

	count := scanner.count()
	// Should continue scanning despite errors
	if count < 2 {
		t.Errorf("scan count = %v, want >= 2 (should continue despite errors)", count)
	}
}

func TestMonitor_ContextCancellation(t *testing.T) {
	scanner := &mockScanner{}
	m := New([]Scanner{scanner}, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)

	count1 := scanner.count()

	// Wait more - count should not increase significantly
	time.Sleep(100 * time.Millisecond)

	count2 := scanner.count()

	// Count should be stable after context cancellation
	if count2 > count1+1 {
		t.Errorf("scan continued after context cancel: count1=%v, count2=%v", count1, count2)
	}
}

func TestMonitor_MultipleScanners(t *testing.T) {
	scanner1 := &mockScanner{}
	scanner2 := &mockScanner{}
	scanner3 := &mockScanner{err: errors.New("always fails")}

	m := New([]Scanner{scanner1, scanner2, scanner3}, 50*time.Millisecond)

	ctx := context.Background()
	m.Start(ctx)

	time.Sleep(150 * time.Millisecond)

	m.Stop()

	// All scanners should be called
	if scanner1.count() < 2 {
		t.Errorf("scanner1 count = %v, want >= 2", scanner1.count())
	}
	if scanner2.count() < 2 {
		t.Errorf("scanner2 count = %v, want >= 2", scanner2.count())
	}
	if scanner3.count() < 2 {
		t.Errorf("scanner3 count = %v, want >= 2 (should be called despite errors)", scanner3.count())
	}
}

func TestMonitor_StopWithoutStart(t *testing.T) {
	m := New([]Scanner{&mockScanner{}}, time.Second)
	// Should not panic
	m.Stop()
}

func TestMonitor_ImmediateScan(t *testing.T) {
	scanner := &mockScanner{}
	m := New([]Scanner{scanner}, 1*time.Hour) // Long interval

	ctx := context.Background()
	m.Start(ctx)

	// Give it a moment to perform immediate scan
	time.Sleep(50 * time.Millisecond)

	m.Stop()

	count := scanner.count()
	// Should have scanned immediately on start
	if count < 1 {
		t.Errorf("scan count = %v, want >= 1 (immediate scan on start)", count)
	}
}
