package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/zfs"
)

// mockDiskRepo is a minimal mock for testing.
type mockDiskRepo struct {
	disks      []store.DiskState
	saveErr    error
	listErr    error
	markDetachErr error
}

func (m *mockDiskRepo) ListAttached() ([]store.DiskState, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.disks, nil
}

func (m *mockDiskRepo) Save(d disk.Info) error {
	return m.saveErr
}

func (m *mockDiskRepo) MarkDetached(name, serial string) error {
	return m.markDetachErr
}

func (m *mockDiskRepo) DeleteSmart(name string) error {
	return nil
}

func (m *mockDiskRepo) SaveSmart(report *disk.DetailedReport) error {
	return nil
}

// mockDiskManager is a minimal mock for testing.
type mockDiskManager struct {
	disks       []disk.Info
	listErr     error
	smartReport *disk.DetailedReport
	smartErr    error
}

func (m *mockDiskManager) ListBasic(ctx context.Context) ([]disk.Info, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.disks, nil
}

func (m *mockDiskManager) SmartDetails(ctx context.Context, name string) (*disk.DetailedReport, error) {
	if m.smartErr != nil {
		return nil, m.smartErr
	}
	return m.smartReport, nil
}

// mockZFSManager is a minimal mock for testing.
type mockZFSManager struct {
	pools   []zfs.Pool
	listErr error
}

func (m *mockZFSManager) ListPools(ctx context.Context) ([]zfs.Pool, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.pools, nil
}

func TestDiskScanner_Scan(t *testing.T) {
	tests := []struct {
		name        string
		current     []disk.Info
		known       []store.DiskState
		listErr     error
		wantEvents  int
		wantErr     bool
	}{
		{
			name: "no_changes",
			current: []disk.Info{
				{Name: "sda", Serial: "SER001"},
			},
			known: []store.DiskState{
				{Name: "sda", Serial: "SER001"},
			},
			wantEvents: 0,
			wantErr:    false,
		},
		{
			name: "disk_added",
			current: []disk.Info{
				{Name: "sda", Serial: "SER001"},
				{Name: "sdb", Serial: "SER002"},
			},
			known: []store.DiskState{
				{Name: "sda", Serial: "SER001"},
			},
			wantEvents: 1, // DiskAdded event
			wantErr:    false,
		},
		{
			name: "disk_removed",
			current: []disk.Info{
				{Name: "sda", Serial: "SER001"},
			},
			known: []store.DiskState{
				{Name: "sda", Serial: "SER001"},
				{Name: "sdb", Serial: "SER002"},
			},
			wantEvents: 1, // DiskRemoved event
			wantErr:    false,
		},
		{
			name:       "list_error",
			current:    []disk.Info{},
			known:      []store.DiskState{},
			listErr:    errors.New("disk list failed"),
			wantEvents: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := event.NewBus()
			repo := &mockDiskRepo{disks: tt.known}
			diskMgr := &mockDiskManager{disks: tt.current, listErr: tt.listErr}

			scanner := NewDiskScanner(bus, repo, diskMgr)

			// Subscribe to count events
			eventCount := 0
			bus.Subscribe(func(e event.Event) {
				eventCount++
			})

			ctx := context.Background()
			err := scanner.Scan(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Give events time to be published
			time.Sleep(10 * time.Millisecond)

			if eventCount != tt.wantEvents {
				t.Errorf("event count = %v, want %v", eventCount, tt.wantEvents)
			}
		})
	}
}

func TestSmartScanner_Scan(t *testing.T) {
	tests := []struct {
		name         string
		disks        []disk.Info
		interval     time.Duration
		waitBefore   time.Duration
		smartReport  *disk.DetailedReport
		wantCollect  bool
	}{
		{
			name: "initial_scan",
			disks: []disk.Info{
				{Name: "sda", Serial: "SER001"},
			},
			interval: time.Hour,
			smartReport: &disk.DetailedReport{
				Disk:   "sda",
				Passed: true,
			},
			wantCollect: true,
		},
		{
			name: "skip_too_soon",
			disks: []disk.Info{
				{Name: "sda", Serial: "SER001"},
			},
			interval:    time.Hour,
			waitBefore:  10 * time.Millisecond, // Less than interval
			wantCollect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := event.NewBus()
			repo := &mockDiskRepo{}
			diskMgr := &mockDiskManager{
				disks:       tt.disks,
				smartReport: tt.smartReport,
			}

			scanner := NewSmartScanner(bus, repo, diskMgr, tt.interval)

			// Set lastUpdate if testing skip
			if tt.waitBefore > 0 {
				scanner.lastUpdate = time.Now()
				time.Sleep(tt.waitBefore)
			}

			ctx := context.Background()
			err := scanner.Scan(ctx)

			if err != nil {
				t.Errorf("Scan() error = %v", err)
			}

			// If we expect collection, lastUpdate should be recent
			if tt.wantCollect && time.Since(scanner.lastUpdate) > time.Second {
				t.Error("SMART data should have been collected but wasn't")
			}
		})
	}
}

func TestSmartScanner_FailedSmartEvent(t *testing.T) {
	bus := event.NewBus()
	repo := &mockDiskRepo{}
	diskMgr := &mockDiskManager{
		disks: []disk.Info{
			{Name: "sda", Serial: "SER001"},
		},
		smartReport: &disk.DetailedReport{
			Disk:   "sda",
			Passed: false, // Failed SMART
		},
	}

	scanner := NewSmartScanner(bus, repo, diskMgr, time.Hour)

	// Subscribe to count SmartFailed events
	eventCount := 0
	bus.Subscribe(func(e event.Event) {
		if e.Type == event.SmartFailed {
			eventCount++
		}
	})

	ctx := context.Background()
	err := scanner.Scan(ctx)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	// Give events time to be published
	time.Sleep(10 * time.Millisecond)

	if eventCount != 1 {
		t.Errorf("SmartFailed event count = %v, want 1", eventCount)
	}
}

func TestZFSScanner_Scan(t *testing.T) {
	tests := []struct {
		name       string
		pools      []zfs.Pool
		listErr    error
		wantEvents int
		wantErr    bool
	}{
		{
			name: "all_healthy",
			pools: []zfs.Pool{
				{Name: "tank", Health: zfs.PoolOnline},
			},
			wantEvents: 0,
			wantErr:    false,
		},
		{
			name: "degraded_pool",
			pools: []zfs.Pool{
				{Name: "tank", Health: zfs.PoolDegraded},
			},
			wantEvents: 1,
			wantErr:    false,
		},
		{
			name: "multiple_pools_mixed",
			pools: []zfs.Pool{
				{Name: "tank1", Health: zfs.PoolOnline},
				{Name: "tank2", Health: zfs.PoolDegraded},
				{Name: "tank3", Health: zfs.PoolFaulted},
			},
			wantEvents: 2, // degraded + faulted
			wantErr:    false,
		},
		{
			name:       "list_error",
			pools:      nil,
			listErr:    errors.New("zfs list failed"),
			wantEvents: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := event.NewBus()
			zfsMgr := &mockZFSManager{
				pools:   tt.pools,
				listErr: tt.listErr,
			}

			scanner := NewZFSScanner(bus, zfsMgr)

			// Subscribe to count events
			eventCount := 0
			bus.Subscribe(func(e event.Event) {
				if e.Type == event.PoolDegraded {
					eventCount++
				}
			})

			ctx := context.Background()
			err := scanner.Scan(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Give events time to be published
			time.Sleep(10 * time.Millisecond)

			if eventCount != tt.wantEvents {
				t.Errorf("event count = %v, want %v", eventCount, tt.wantEvents)
			}
		})
	}
}
