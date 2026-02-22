package task

import (
    "context"
    "sync"
    "testing"
    "time"

    "github.com/google/uuid"
)

// TestManagerSubmit tests task submission and execution
func TestManagerSubmit(t *testing.T) {
    tests := []struct {
        name      string
        taskName  string
        taskFunc  func(ctx context.Context, update func(progress int)) (interface{}, error)
        wantState State
    }{
        {
            name:     "successful task",
            taskName: "test-task",
            taskFunc: func(ctx context.Context, update func(progress int)) (interface{}, error) {
                update(50)
                return "result", nil
            },
            wantState: StateDone,
        },
        {
            name:     "task with error",
            taskName: "error-task",
            taskFunc: func(ctx context.Context, update func(progress int)) (interface{}, error) {
                update(25)
                return nil, context.DeadlineExceeded
            },
            wantState: StateFailed,
        },
        {
            name:     "cancelled task",
            taskName: "cancel-task",
            taskFunc: func(ctx context.Context, update func(progress int)) (interface{}, error) {
                <-ctx.Done()
                return nil, ctx.Err()
            },
            wantState: StateCancelled,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            op, err := manager.Submit(tt.taskName, tt.taskFunc)
            if err != nil {
                t.Fatalf("Submit() returned error: %v", err)
            }

            // Wait for task completion
            time.Sleep(100 * time.Millisecond)

            // Verify operation state
            if op.State != tt.wantState {
                t.Errorf("got state %v, want %v", op.State, tt.wantState)
            }

            if op.Name != tt.taskName {
                t.Errorf("got name %v, want %v", op.Name, tt.taskName)
            }

            if op.State == StateRunning {
                t.Error("task should not be running after completion")
            }
        })
    }
}

// TestManagerGet tests task retrieval
func TestManagerGet(t *testing.T) {
    tests := []struct {
        name     string
        exists   bool
        setup    func(m *Manager) string
    }{
        {
            name: "existing task",
            exists: true,
            setup: func(m *Manager) string {
                id := uuid.New().String()
                m.tasks[id] = &Operation{
                    ID:    id,
                    Name:  "test",
                    State: StateRunning,
                }
                return id
            },
        },
        {
            name:   "non-existing task",
            exists: false,
            setup: func(m *Manager) string {
                return uuid.New().String()
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := tt.setup(manager)
            op, exists := manager.Get(id)

            if tt.exists && !exists {
                t.Error("expected task to exist")
            }
            if !tt.exists && exists {
                t.Error("expected task to not exist")
            }
            if tt.exists && op != nil {
                if op.ID != id {
                    t.Errorf("got id %v, want %v", op.ID, id)
                }
            }
        })
    }
}

// TestManagerList tests task listing
func TestManagerList(t *testing.T) {
    tests := []struct {
        name     string
        tasks    []*Operation
        limit    int
        offset   int
        wantLen  int
    }{
        {
            name:    "empty list",
            tasks:   []*Operation{},
            limit:   10,
            offset:  0,
            wantLen: 0,
        },
        {
            name: "full list",
            tasks: []*Operation{
                {Name: "task1"},
                {Name: "task2"},
                {Name: "task3"},
            },
            limit:   10,
            offset:  0,
            wantLen: 3,
        },
        {
            name: "paginated list",
            tasks: []*Operation{
                {Name: "task1"},
                {Name: "task2"},
                {Name: "task3"},
            },
            limit:   2,
            offset:  0,
            wantLen: 2,
        },
        {
            name: "offset pagination",
            tasks: []*Operation{
                {Name: "task1"},
                {Name: "task2"},
                {Name: "task3"},
            },
            limit:   2,
            offset:  1,
            wantLen: 2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            // Populate manager with test tasks
            for _, op := range tt.tasks {
                id := uuid.New().String()
                op.ID = id
                manager.tasks[id] = op
            }

            got, err := manager.List(tt.limit, tt.offset)
            if err != nil {
                t.Fatalf("List() returned error: %v", err)
            }

            if len(got) != tt.wantLen {
                t.Errorf("got %d tasks, want %d", len(got), tt.wantLen)
            }
        })
    }
}

// TestManagerCancel tests task cancellation
func TestManagerCancel(t *testing.T) {
    tests := []struct {
        name   string
        setup  func(m *Manager) string
        cancel bool
        want   State
    }{
        {
            name: "cancel running task",
            setup: func(m *Manager) string {
                id := uuid.New().String()
                ctx, cancel := context.WithCancel(context.Background())
                m.tasks[id] = &Operation{
                    ID:       id,
                    Name:     "running",
                    State:    StateRunning,
                    cancelFn: cancel,
                }
                return id
            },
            cancel: true,
            want:   StateCancelled,
        },
        {
            name: "cancel non-running task",
            setup: func(m *Manager) string {
                id := uuid.New().String()
                m.tasks[id] = &Operation{
                    ID:    id,
                    Name:  "done",
                    State: StateDone,
                }
                return id
            },
            cancel: true,
            want:   StateDone, // Should remain done
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := tt.setup(manager)
            if tt.cancel {
                manager.Cancel(id)
            }

            op, exists := manager.Get(id)
            if !exists {
                t.Error("task should exist")
            }
            if exists && op.State != tt.want {
                t.Errorf("got state %v, want %v", op.State, tt.want)
            }
        })
    }
}

// TestOperationCancel tests the Operation Cancel method
func TestOperationCancel(t *testing.T) {
    tests := []struct {
        name      string
        cancelFn  context.CancelFunc
        wantPanic bool
    }{
        {
            name:      "valid cancel function",
            cancelFn:  func() { cancel() },
            wantPanic: false,
        },
        {
            name:      "nil cancel function",
            cancelFn:  nil,
            wantPanic: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            op := &Operation{
                cancelFn: tt.cancelFn,
            }

            defer func() {
                r := recover()
                if tt.wantPanic && r == nil {
                    t.Error("expected panic but got none")
                }
                if !tt.wantPanic && r != nil {
                    t.Errorf("got unexpected panic: %v", r)
                }
            }()

            // Create a dummy cancel function for the first test
            if tt.cancelFn != nil {
                ctx, cancel := context.WithCancel(context.Background())
                op.cancelFn = cancel
                defer ctx.Done()
            }

            op.Cancel()
        })
    }
}

// TestConcurrency tests concurrent task operations
func TestConcurrency(t *testing.T) {
    manager := &Manager{
        tasks: make(map[string]*Operation),
    }

    var wg sync.WaitGroup
    const numTasks = 10

    for i := 0; i < numTasks; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            taskFunc := func(ctx context.Context, update func(progress int)) (interface{}, error) {
                update(50)
                return i, nil
            }
            _, err := manager.Submit("task", taskFunc)
            if err != nil {
                t.Errorf("Submit() returned error: %v", err)
            }
        }(i)
    }

    wg.Wait()
    // Give tasks time to complete
    time.Sleep(200 * time.Millisecond)

    // Verify all tasks completed
    ops, err := manager.List(100, 0)
    if err != nil {
        t.Fatalf("List() returned error: %v", err)
    }

    if len(ops) != numTasks {
        t.Errorf("got %d tasks, want %d", len(ops), numTasks)
    }

    // Verify no tasks are still running
    for _, op := range ops {
        if op.State == StateRunning {
            t.Error("task should not be running after completion")
        }
    }
}

// TestManagerClose tests proper shutdown
func TestManagerClose(t *testing.T) {
    manager := &Manager{
        tasks: make(map[string]*Operation),
        wg:    sync.WaitGroup{},
    }

    // Start a long-running task
    ctx, cancel := context.WithCancel(context.Background())
    manager.wg.Add(1)
    go func() {
        defer manager.wg.Done()
        <-ctx.Done()
    }()

    // Should not hang
    done := make(chan struct{})
    go func() {
        defer close(done)
        manager.Close()
    }()

    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Error("Close() timed out")
    }

    cancel()
}

// TestUpdateProgress tests progress updates
func TestUpdateProgress(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        updated  int
        want     int
    }{
        {
            name:    "update from 0 to 50",
            initial: 0,
            updated: 50,
            want:    50,
        },
        {
            name:    "update from 50 to 100",
            initial: 50,
            updated: 100,
            want:    100,
        },
        {
            name:    "invalid progress",
            initial: 0,
            updated: 150,
            want:    150, // Manager doesn't validate range
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := uuid.New().String()
            manager.tasks[id] = &Operation{
                ID:      id,
                State:   StateRunning,
                Progress: tt.initial,
            }

            manager.updateProgressNoError(id, tt.updated)

            op := manager.tasks[id]
            if op.Progress != tt.want {
                t.Errorf("got progress %d, want %d", op.Progress, tt.want)
            }
        })
    }
}

// TestUpdateState tests state transitions
func TestUpdateState(t *testing.T) {
    tests := []struct {
        name   string
        from   State
        to     State
        errors bool
    }{
        {
            name: "pending to running",
            from: StatePending,
            to:   StateRunning,
        },
        {
            name: "running to done",
            from: StateRunning,
            to:   StateDone,
        },
        {
            name: "running to failed",
            from: StateRunning,
            to:   StateFailed,
        },
        {
            name: "running to cancelled",
            from: StateRunning,
            to:   StateCancelled,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := uuid.New().String()
            manager.tasks[id] = &Operation{
                ID:    id,
                State: tt.from,
            }

            manager.updateState(id, tt.to, 100, "result", nil)

            op := manager.tasks[id]
            if op.State != tt.to {
                t.Errorf("got state %v, want %v", op.State, tt.to)
            }
        })
    }
}

// TestManagerWait tests waiting for task completion
func TestManagerWait(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(m *Manager) string
        waitTime  time.Duration
        wantError bool
    }{
        {
            name: "wait for task completion",
            setup: func(m *Manager) string {
                taskFunc := func(ctx context.Context, update func(progress int)) (interface{}, error) {
                    time.Sleep(50 * time.Millisecond)
                    return "completed", nil
                }
                op, _ := m.Submit("completion-test", taskFunc)
                return op.ID
            },
            waitTime:  200 * time.Millisecond,
            wantError: false,
        },
        {
            name: "wait for non-existent task",
            setup: func(m *Manager) string {
                return uuid.New().String()
            },
            waitTime:  0,
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := tt.setup(manager)
            if tt.waitTime > 0 {
                time.Sleep(tt.waitTime)
            }

            _, err := manager.Wait(id)
            if tt.wantError && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.wantError && err != nil {
                t.Errorf("got unexpected error: %v", err)
            }
        })
    }
}

// TestManagerCount tests counting active tasks
func TestManagerCount(t *testing.T) {
    manager := &Manager{
        tasks: make(map[string]*Operation),
    }

    // Initially empty
    if count := manager.Count(); count != 0 {
        t.Errorf("got count %d, want 0", count)
    }

    // Add running task
    id1 := uuid.New().String()
    manager.tasks[id1] = &Operation{
        ID:    id1,
        State: StateRunning,
    }

    if count := manager.Count(); count != 1 {
        t.Errorf("got count %d, want 1", count)
    }

    // Add another task
    id2 := uuid.New().String()
    manager.tasks[id2] = &Operation{
        ID:    id2,
        State: StateDone, // Completed task still counts as active until deleted
    }

    if count := manager.Count(); count != 2 {
        t.Errorf("got count %d, want 2", count)
    }
}

// TestManagerListActive tests listing active tasks
func TestManagerListActive(t *testing.T) {
    manager := &Manager{
        tasks: make(map[string]*Operation),
    }

    // Add tasks with different states
    pendingID := uuid.New().String()
    runningID := uuid.New().String()
    doneID := uuid.New().String()
    failedID := uuid.New().String()

    manager.tasks[pendingID] = &Operation{
        ID:    pendingID,
        Name:  "pending",
        State: StatePending,
    }
    manager.tasks[runningID] = &Operation{
        ID:    runningID,
        Name:  "running",
        State: StateRunning,
    }
    manager.tasks[doneID] = &Operation{
        ID:    doneID,
        Name:  "done",
        State: StateDone,
    }
    manager.tasks[failedID] = &Operation{
        ID:    failedID,
        Name:  "failed",
        State: StateFailed,
    }

    active := manager.ListActive()
    if len(active) != 2 {
        t.Errorf("got %d active tasks, want 2", len(active))
    }

    // Verify correct tasks are active
    activeIDs := make(map[string]bool)
    for _, op := range active {
        activeIDs[op.ID] = true
    }

    if !activeIDs[pendingID] {
        t.Error("pending task should be active")
    }
    if !activeIDs[runningID] {
        t.Error("running task should be active")
    }
    if activeIDs[doneID] {
        t.Error("done task should not be active")
    }
    if activeIDs[failedID] {
        t.Error("failed task should not be active")
    }
}

// TestManagerDelete tests deleting tasks
func TestManagerDelete(t *testing.T) {
    tests := []struct {
        name       string
        exists     bool
        setup      func(m *Manager) string
        wantError  bool
    }{
        {
            name:      "delete existing task",
            exists:    true,
            wantError: false,
            setup: func(m *Manager) string {
                id := uuid.New().String()
                m.tasks[id] = &Operation{ID: id}
                return id
            },
        },
        {
            name:      "delete non-existent task",
            exists:    false,
            wantError: true,
            setup: func(m *Manager) string {
                return uuid.New().String()
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := &Manager{
                tasks: make(map[string]*Operation),
            }

            id := tt.setup(manager)
            err := manager.Delete(id)
            if tt.wantError && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.wantError && err != nil {
                t.Errorf("got unexpected error: %v", err)
            }

            // Verify task was deleted
            if !tt.wantError {
                _, exists := manager.Get(id)
                if exists {
                    t.Error("task should have been deleted")
                }
            }
        })
    }
}