package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

// mockFetcher simulates the process fetcher for testing
type mockFetcher struct {
	processes []*process.Process
	err       error
}

func (m *mockFetcher) Processes(ctx context.Context) ([]*process.Process, error) {
	return m.processes, m.err
}

func TestSystemMonitor_UpdateCache(t *testing.T) {
	// Create mock processes
	// Note: We can't easily create process.Process structs with internal state populated
	// because fields are private. However, we can verify the caching mechanism itself.

	p1 := &process.Process{Pid: 1}
	p2 := &process.Process{Pid: 2}

	tests := []struct {
		name          string
		initialCache  map[int32]*process.Process
		fetchedProcs  []*process.Process
		fetchedErr    error
		expectedCache map[int32]*process.Process
	}{
		{
			name:          "First run - populates cache",
			initialCache:  map[int32]*process.Process{},
			fetchedProcs:  []*process.Process{p1, p2},
			fetchedErr:    nil,
			expectedCache: map[int32]*process.Process{1: p1, 2: p2},
		},
		{
			name:          "Update - preserves existing objects",
			initialCache:  map[int32]*process.Process{1: p1},
			fetchedProcs:  []*process.Process{p1, p2},
			fetchedErr:    nil,
			expectedCache: map[int32]*process.Process{1: p1, 2: p2},
		},
		{
			name:          "Update - removes dead processes",
			initialCache:  map[int32]*process.Process{1: p1, 2: p2},
			fetchedProcs:  []*process.Process{p1},
			fetchedErr:    nil,
			expectedCache: map[int32]*process.Process{1: p1},
		},
		{
			name:          "Error fetching - keeps old cache (safety)",
			initialCache:  map[int32]*process.Process{1: p1},
			fetchedProcs:  nil,
			fetchedErr:    assert.AnError,
			expectedCache: map[int32]*process.Process{1: p1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewSystemMonitor()
			m.fetcher = &mockFetcher{processes: tc.fetchedProcs, err: tc.fetchedErr}

			// Setup initial state
			m.mu.Lock()
			m.cache = tc.initialCache
			m.mu.Unlock()

			// Act
			m.updateCache(context.Background())

			// Assert
			m.mu.RLock()
			defer m.mu.RUnlock()

			if tc.fetchedErr != nil {
				// On error, cache should not change (actually updateCache returns early)
				// My implementation returns early, so cache remains same.
				// However, if we didn't mock the mutex, we'd need to be careful.
				// In this test, we just check if it matches expectation.
				assert.Equal(t, len(tc.initialCache), len(m.cache))
			} else {
				assert.Equal(t, len(tc.expectedCache), len(m.cache))
				for pid, expectedProc := range tc.expectedCache {
					actualProc, ok := m.cache[pid]
					assert.True(t, ok)
					// Verify pointer identity to ensure we kept the old object
					if pid == 1 && tc.initialCache[1] != nil {
						assert.True(t, expectedProc == actualProc, "Should preserve pointer identity")
					}
				}
			}
		})
	}
}

func TestSystemMonitor_GetProcesses(t *testing.T) {
	m := NewSystemMonitor()

	// Create a dummy process
	// We can't fully mock process.Process behavior for CPUPercent without mocking the internal OS calls
	// or using the internal testing hooks of gopsutil.
	// So we test that GetProcesses returns what's in the cache transformed to ProcessInfo.

	p1 := &process.Process{Pid: 100}

	m.mu.Lock()
	m.cache = map[int32]*process.Process{100: p1}
	m.mu.Unlock()

	ctx := context.Background()
	infos, err := m.GetProcesses(ctx)

	assert.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, int32(100), infos[0].PID)
}

func TestSystemMonitor_Lifecycle(t *testing.T) {
	m := NewSystemMonitor()

	// Mock fetcher to avoid real OS calls
	m.fetcher = &mockFetcher{processes: []*process.Process{}, err: nil}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)

	// Give it a moment to run
	time.Sleep(100 * time.Millisecond)

	m.Stop()

	// If we reach here without panic/deadlock, it's good.
	// We could inspect internal state if needed.
}
