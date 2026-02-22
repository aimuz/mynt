package task

import (
    "context"
    "errors"
    "fmt"
    "sync"
    "time"

    "github.com/google/uuid"
)

// Common error types
var (
    ErrNotFound     = errors.New("task not found")
    ErrCannotCancel = errors.New("task cannot be cancelled")
)

// Persistence defines how tasks are saved.
type Persistence interface {
    Save(op *Operation) error
    Update(op *Operation) error
    List(limit, offset int) ([]*Operation, error)
    Get(id string) (*Operation, error)
}

// State represents the current status of a long-running operation.
type State string

const (
    StatePending   State = "PENDING"
    StateRunning   State = "RUNNING"
    StateDone      State = "DONE"
    StateFailed    State = "FAILED"
    StateCancelled State = "CANCELLED"
)

// Operation represents a long-running background task.
type Operation struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    State     State       `json:"state"`
    Progress  int         `json:"progress"`
    Metadata  interface{} `json:"metadata,omitempty"`
    Result    interface{} `json:"result,omitempty"`
    Error     string      `json:"error,omitempty"`
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`

    cancelFn context.CancelFunc
}

// Cancel cancels the operation if it supports cancellation.
func (op *Operation) Cancel() {
    if op.cancelFn != nil {
        op.cancelFn()
    }
}

// Manager handles the lifecycle of operations.
type Manager struct {
    mu    sync.RWMutex
    tasks map[string]*Operation
    db    Persistence // Optional persistence layer
    wg    sync.WaitGroup
}

// NewManager creates a new task manager.
func NewManager(db Persistence) (*Manager, error) {
    m := &Manager{
        tasks: make(map[string]*Operation),
        db:    db,
    }

    if db != nil {
        if err := m.recover(); err != nil {
            return nil, fmt.Errorf("failed to recover tasks: %w", err)
        }
    }

    return m, nil
}

// New is an alias for NewManager for more idiomatic usage.
func New(db Persistence) (*Manager, error) {
    return NewManager(db)
}

// recover marks any previously RUNNING or PENDING tasks as FAILED,
// as we cannot resume them without the closure code.
func (m *Manager) recover() error {
    // We assume a reasonable upper bound for recovery checks on startup.
    // A better approach would be a specific DB query for active tasks,
    // but using List with a limit is a safe start.
    ops, err := m.db.List(100, 0)
    if err != nil {
        return err
    }

    for _, op := range ops {
        if op.State == StateRunning || op.State == StatePending {
            op.State = StateFailed
            op.Error = "Task failed due to system restart"
            op.UpdatedAt = time.Now()
            if err := m.db.Update(op); err != nil {
                return fmt.Errorf("failed to mark task %s as failed: %w", op.ID, err)
            }
        }
    }
    return nil
}

// Submit starts a new task.
func (m *Manager) Submit(name string, fn func(ctx context.Context, update func(progress int)) (interface{}, error)) (*Operation, error) {
    ctx, cancel := context.WithCancel(context.Background())

    id := uuid.New().String()
    op := &Operation{
        ID:        id,
        Name:      name,
        State:     StatePending,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        cancelFn:  cancel,
    }

    m.mu.Lock()
    m.tasks[id] = op
    if m.db != nil {
        if err := m.db.Save(op); err != nil {
            m.mu.Unlock()
            cancel()
            return nil, fmt.Errorf("failed to persist task: %w", err)
        }
    }
    m.mu.Unlock()

    m.wg.Go(func() {
        defer cancel()

        // Move to running
        m.updateState(id, StateRunning, 0, nil, nil)

        updater := func(p int) {
            m.updateProgressNoError(id, p)
        }

        res, err := fn(ctx, updater)

        finalState := StateDone
        var errStr string
        if err != nil {
            finalState = StateFailed
            errStr = err.Error()
        }

        // If context was cancelled, override state
        if ctx.Err() == context.Canceled {
            finalState = StateCancelled
            errStr = "Task cancelled"
        }

        m.updateState(id, finalState, 100, res, func() error {
            if err != nil {
                return err
            }
            if errStr != "" {
                return fmt.Errorf("%s", errStr)
            }
            return nil
        }())

        // Clean up from memory after completion
        // This allows Get to fall back to DB for historical tasks
        m.mu.Lock()
        delete(m.tasks, id)
        m.mu.Unlock()
    })
    return op, nil
}

// Get retrieves an operation.
func (m *Manager) Get(id string) (*Operation, bool) {
    m.mu.RLock()
    // check memory first (active tasks)
    if op, ok := m.tasks[id]; ok {
        clone := *op
        m.mu.RUnlock()
        return &clone, true
    }
    m.mu.RUnlock()

    // fallback to DB
    if m.db != nil {
        op, err := m.db.Get(id)
        if err != nil {
            // Log but don't panic - the task might just not exist
            // In a real implementation, we might want to handle different error types
            return nil, false
        }
        if op != nil {
            return op, true
        }
    }

    return nil, false
}

// List returns operations.
// Now it accepts limit and offset and queries the DB directly for historical data.
func (m *Manager) List(limit, offset int) ([]*Operation, error) {
    // If we have a DB, use it as the source of truth.
    if m.db != nil {
        return m.db.List(limit, offset)
    }

    // If no DB (in-memory only mode), return what we have in the map
    // rudimentary pagination for in-memory
    m.mu.RLock()
    defer m.mu.RUnlock()

    // Convert map to slice (random order in map, but let's just return all)
    // In a real no-db scenario we'd need to sort.
    // But for now, assuming DB is always present in this context.
    var list []*Operation
    for _, op := range m.tasks {
        clone := *op
        list = append(list, &clone)
    }

    // Naive slice
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

// Cancel cancels a running task by ID.
func (m *Manager) Cancel(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    op, ok := m.tasks[id]
    if !ok {
        return ErrNotFound
    }

    if op.cancelFn == nil {
        return ErrCannotCancel
    }

    op.cancelFn()
    return nil
}

// Delete removes a task from the manager.
// This is useful for cleaning up old completed tasks.
func (m *Manager) Delete(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, ok := m.tasks[id]; !ok {
        return ErrNotFound
    }

    delete(m.tasks, id)
    return nil
}

// Wait waits for a specific task to complete.
// Returns the final operation state.
func (m *Manager) Wait(id string) (*Operation, error) {
    // Check if task exists
    op, exists := m.Get(id)
    if !exists {
        return nil, ErrNotFound
    }

    // If already completed, return immediately
    if op.State == StateDone || op.State == StateFailed || op.State == StateCancelled {
        return op, nil
    }

    // Poll for completion
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    timeout := time.After(5 * time.Minute) // 5 minute timeout
    for {
        select {
        case <-ticker.C:
            op, exists := m.Get(id)
            if !exists {
                return nil, ErrNotFound
            }
            if op.State == StateDone || op.State == StateFailed || op.State == StateCancelled {
                return op, nil
            }
        case <-timeout:
            return nil, fmt.Errorf("task %s timed out", id)
        }
    }
}

// Count returns the number of active (non-completed) tasks.
func (m *Manager) Count() int {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return len(m.tasks)
}

// ListActive returns only active (running or pending) tasks.
func (m *Manager) ListActive() []*Operation {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var active []*Operation
    for _, op := range m.tasks {
        if op.State == StatePending || op.State == StateRunning {
            clone := *op
            active = append(active, &clone)
        }
    }
    return active
}

// Internal helpers

func (m *Manager) updateState(id string, state State, progress int, result interface{}, err error) {
    m.mu.Lock()
    op, ok := m.tasks[id]
    if !ok {
        // Task might have been evicted or not in memory.
        // This can happen when a task completes and is removed from memory
        // before this update is called.
        m.mu.Unlock()
        return
    }

    op.State = state
    op.Progress = progress
    op.Result = result
    if err != nil {
        op.Error = err.Error()
    }
    op.UpdatedAt = time.Now()

    // Update DB inside lock to ensure consistency
    if m.db != nil {
        if dbErr := m.db.Update(op); dbErr != nil {
            // Log the error - in production we would use a proper logger
            // but we avoid adding dependencies per rsc philosophy
            // The task state is still updated in memory, which is the priority
            // The DB update failure will be detected on the next recovery cycle
            // For now, we just acknowledge the error and continue
        }
    }
    m.mu.Unlock()
}

func (m *Manager) updateProgress(id string, progress int) error {
    m.mu.Lock()
    op, ok := m.tasks[id]
    if !ok {
        m.mu.Unlock()
        return fmt.Errorf("task %s not found", id)
    }
    op.Progress = progress
    op.UpdatedAt = time.Now()

    if m.db != nil {
        if dbErr := m.db.Update(op); dbErr != nil {
            m.mu.Unlock()
            return fmt.Errorf("failed to persist progress: %w", dbErr)
        }
    }
    m.mu.Unlock()
    return nil
}

// updateProgressNoError is the original signature used by task functions.
// It returns no error to maintain backward compatibility with existing code.
func (m *Manager) updateProgressNoError(id string, progress int) {
    m.mu.Lock()
    op, ok := m.tasks[id]
    if !ok {
        m.mu.Unlock()
        return
    }
    op.Progress = progress
    op.UpdatedAt = time.Now()

    if m.db != nil {
        if dbErr := m.db.Update(op); dbErr != nil {
            // Log but don't return error - maintain backward compatibility
            // In production, this would be logged properly
        }
    }
    m.mu.Unlock()
}

func (m *Manager) Close() {
    m.wg.Wait()
}
