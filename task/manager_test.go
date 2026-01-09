package task

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockPersistence is a simple in-memory persistence for testing.
type mockPersistence struct {
	mu   sync.RWMutex
	ops  map[string]*Operation
	errs map[string]error
}

func newMockPersistence() *mockPersistence {
	return &mockPersistence{
		ops:  make(map[string]*Operation),
		errs: make(map[string]error),
	}
}

func (m *mockPersistence) Save(op *Operation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.errs["Save"]; err != nil {
		return err
	}
	clone := *op
	m.ops[op.ID] = &clone
	return nil
}

func (m *mockPersistence) Update(op *Operation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.errs["Update"]; err != nil {
		return err
	}
	if _, exists := m.ops[op.ID]; !exists {
		return errors.New("operation not found")
	}
	clone := *op
	m.ops[op.ID] = &clone
	return nil
}

func (m *mockPersistence) List(limit, offset int) ([]*Operation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.errs["List"]; err != nil {
		return nil, err
	}
	var list []*Operation
	for _, op := range m.ops {
		clone := *op
		list = append(list, &clone)
	}
	start := offset
	if start > len(list) {
		start = len(list)
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], nil
}

func (m *mockPersistence) Get(id string) (*Operation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.errs["Get"]; err != nil {
		return nil, err
	}
	if op, ok := m.ops[id]; ok {
		clone := *op
		return &clone, nil
	}
	return nil, errors.New("not found")
}

func (m *mockPersistence) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errs[method] = err
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		db      Persistence
		wantErr bool
	}{
		{
			name:    "no_persistence",
			db:      nil,
			wantErr: false,
		},
		{
			name:    "with_persistence",
			db:      newMockPersistence(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && m == nil {
				t.Error("NewManager() returned nil manager")
			}
		})
	}
}

func TestManager_Submit(t *testing.T) {
	tests := []struct {
		name      string
		taskName  string
		fn        func(ctx context.Context, update func(progress int)) (interface{}, error)
		wantState State
		wantErr   bool
	}{
		{
			name:     "success",
			taskName: "test_task",
			fn: func(ctx context.Context, update func(progress int)) (interface{}, error) {
				update(50)
				return "done", nil
			},
			wantState: StateDone,
			wantErr:   false,
		},
		{
			name:     "failure",
			taskName: "failing_task",
			fn: func(ctx context.Context, update func(progress int)) (interface{}, error) {
				return nil, errors.New("task failed")
			},
			wantState: StateFailed,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newMockPersistence()
			m, err := NewManager(db)
			if err != nil {
				t.Fatalf("NewManager() error = %v", err)
			}
			defer m.Close()

			op, err := m.Submit(tt.taskName, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Submit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if op.Name != tt.taskName {
				t.Errorf("Submit() name = %v, want %v", op.Name, tt.taskName)
			}
			if op.State != StatePending {
				t.Errorf("Submit() initial state = %v, want %v", op.State, StatePending)
			}

			// Wait for task to complete
			timeout := time.After(2 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					t.Fatal("task did not complete in time")
				case <-ticker.C:
					retrieved, ok := m.Get(op.ID)
					if !ok {
						// Task completed and was removed from memory
						// Check DB
						dbOp, err := db.Get(op.ID)
						if err != nil {
							continue
						}
						if dbOp.State == tt.wantState {
							return
						}
						t.Errorf("final state = %v, want %v", dbOp.State, tt.wantState)
						return
					}
					if retrieved.State == tt.wantState {
						return
					}
					if retrieved.State == StateFailed && tt.wantState != StateFailed {
						t.Errorf("task failed unexpectedly: %v", retrieved.Error)
						return
					}
				}
			}
		})
	}
}

func TestManager_Get(t *testing.T) {
	db := newMockPersistence()
	m, err := NewManager(db)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer m.Close()

	// Submit a task
	op, err := m.Submit("test", func(ctx context.Context, update func(progress int)) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "done", nil
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	tests := []struct {
		name   string
		id     string
		wantOK bool
	}{
		{
			name:   "existing",
			id:     op.ID,
			wantOK: true,
		},
		{
			name:   "non_existing",
			id:     "nonexistent",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := m.Get(tt.id)
			if ok != tt.wantOK {
				t.Errorf("Get() ok = %v, wantOK %v", ok, tt.wantOK)
			}
			if tt.wantOK && got == nil {
				t.Error("Get() returned nil operation")
			}
		})
	}
}

func TestManager_List(t *testing.T) {
	db := newMockPersistence()
	m, err := NewManager(db)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer m.Close()

	// Submit multiple tasks
	for i := 0; i < 3; i++ {
		_, err := m.Submit("task", func(ctx context.Context, update func(progress int)) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return "done", nil
		})
		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	// Wait a bit for tasks to be saved
	time.Sleep(200 * time.Millisecond)

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all",
			limit:     10,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "limited",
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "offset",
			limit:     10,
			offset:    2,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := m.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(list) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(list), tt.wantCount)
			}
		})
	}
}

func TestManager_Recover(t *testing.T) {
	db := newMockPersistence()

	// Manually add a "stuck" running task
	stuckOp := &Operation{
		ID:        "stuck",
		Name:      "stuck_task",
		State:     StateRunning,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
	if err := db.Save(stuckOp); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create manager - should recover stuck tasks
	m, err := NewManager(db)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer m.Close()

	// Check that stuck task was marked as failed
	recovered, err := db.Get("stuck")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if recovered.State != StateFailed {
		t.Errorf("recovered state = %v, want %v", recovered.State, StateFailed)
	}
	if recovered.Error == "" {
		t.Error("recovered task should have error message")
	}
}

func TestManager_SubmitWithPersistenceError(t *testing.T) {
	db := newMockPersistence()
	db.setError("Save", errors.New("db error"))

	m, err := NewManager(nil) // Start without DB to avoid recovery issues
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	m.db = db // Set DB after creation
	defer m.Close()

	_, err = m.Submit("test", func(ctx context.Context, update func(progress int)) (interface{}, error) {
		return "done", nil
	})
	if err == nil {
		t.Error("Submit() expected error with failed persistence, got nil")
	}
}

func TestManager_InMemoryMode(t *testing.T) {
	// Manager without persistence should work
	m, err := NewManager(nil)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer m.Close()

	op, err := m.Submit("test", func(ctx context.Context, update func(progress int)) (interface{}, error) {
		return "done", nil
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	// Should be retrievable while running
	_, ok := m.Get(op.ID)
	if !ok {
		t.Error("Get() failed to retrieve in-memory task")
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	// List should return tasks (though empty after completion in-memory only mode)
	list, err := m.List(10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	// In-memory mode: completed tasks are removed, so list might be empty
	_ = list
}
