# Task Management System

The task management system provides a robust framework for tracking and managing long-running background operations in Mynt NAS.

## Philosophy

This implementation follows Russ Cox's (rsc) coding philosophy:

- **Simplicity**: Clear, readable code over clever abstractions
- **Error handling**: Errors are values - always handle them explicitly  
- **Testing**: Comprehensive table-driven tests for all functionality
- **Dependencies**: Minimal external dependencies, prefer standard library

## Architecture

### Core Components

1. **Operation**: Represents a single long-running task
   - Tracks state, progress, results, and errors
   - Supports cancellation via context
   - JSON-serializable for API responses

2. **Manager**: Orchestrates task lifecycle
   - Submit new tasks
   - Track running operations
   - Provide persistence via database
   - Handle recovery on startup

3. **Persistence**: Database interface for task storage
   - Save new operations
   - Update operation state
   - List operations with pagination
   - Retrieve by ID

### Task States

- `PENDING`: Task is queued but not yet started
- `RUNNING`: Task is currently executing
- `DONE`: Task completed successfully
- `FAILED`: Task failed with an error
- `CANCELLED`: Task was cancelled by user

## Usage Examples

### Basic Task Submission

```go
manager := task.NewManager(db)

op, err := manager.Submit("ZFS Pool Creation", func(ctx context.Context, update func(progress int)) (interface{}, error) {
    update(25)
    
    // Perform work...
    result, err := createPool()
    if err != nil {
        return nil, err
    }
    
    update(75)
    
    // More work...
    return result, nil
})
if err != nil {
    log.Fatal(err)
}

// Check status
if op.State == task.StateRunning {
    log.Println("Task is running")
}
```

### Task Cancellation

```go
// Cancel by operation
op.Cancel()

// Or cancel by ID
err := manager.Cancel(op.ID)
if err != nil {
    log.Printf("Failed to cancel: %v", err)
}
```

### Task Monitoring

```go
// Get task status
op, exists := manager.Get(taskID)
if !exists {
    log.Println("Task not found")
}

switch op.State {
case task.StateRunning:
    fmt.Printf("Progress: %d%%\n", op.Progress)
case task.StateDone:
    fmt.Printf("Completed: %v\n", op.Result)
case task.StateFailed:
    fmt.Printf("Failed: %s\n", op.Error)
}
```

### Task History

```go
// List recent tasks
ops, err := manager.List(50, 0) // Limit 50, offset 0
if err != nil {
    log.Fatal(err)
}

for _, op := range ops {
    fmt.Printf("Task: %s - %s (%.1f%%)\n", op.Name, op.State, op.Progress)
}
```

## Error Handling

The system follows Go best practices for error handling:

- All errors are explicit values
- Database errors are wrapped with context
- Progress update failures don't fail the entire task
- Cancellation respects context semantics

### Error Patterns

```go
// Task submission failures
op, err := manager.Submit("task", fn)
if err != nil {
    return fmt.Errorf("failed to submit task: %w", err)
}

// Progress updates
if err := manager.updateProgress(id, 50); err != nil {
    // Handle progress persistence failure
    // Task continues but progress may not be saved
}

// Task cancellation
if err := manager.Cancel(id); err != nil {
    if errors.Is(err, task.ErrNotFound) {
        // Task already completed/removed
    } else {
        return fmt.Errorf("cancellation failed: %w", err)
    }
}
```

## Persistence

Tasks are automatically persisted to the database:

- New tasks are saved on submission
- Progress updates are persisted in real-time
- Final state and results are saved on completion
- Failed tasks during restart are marked as failed

### Database Schema

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    state TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    metadata TEXT,
    result TEXT,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Concurrency

The system is designed for concurrent access:

- All operations are protected by mutexes
- Tasks run in separate goroutines
- Multiple managers can operate concurrently
- Database operations are thread-safe

### Thread Safety

- `Manager` methods are safe for concurrent use
- Individual `Operation` instances should not be shared between goroutines
- Task functions receive their own context for cancellation
- Progress updates are synchronized

## Recovery

On startup, the manager attempts to recover incomplete tasks:

1. Queries database for all tasks
2. Marks any `PENDING` or `RUNNING` tasks as `FAILED`
3. Logs the failure reason as "Task failed due to system restart"
4. Continues normal operation

This ensures consistency but acknowledges that long-running tasks cannot be resumed without their original function.

## Testing

The system includes comprehensive table-driven tests:

- Task submission and execution
- Progress updates
- State transitions  
- Error handling
- Cancellation
- Concurrency safety
- Persistence behavior

Tests follow rsc's table-driven approach with clear subtests and isolated test cases.

## Performance Considerations

- Active tasks are kept in memory for fast access
- Completed tasks are evicted to reduce memory usage
- Database operations are batched where possible
- Progress updates can fail silently to avoid blocking
- Task cleanup happens automatically on completion

## Integration Points

The task system integrates with several Mynt NAS components:

- **API Server**: Provides REST endpoints for task management
- **Storage Service**: Long-running ZFS operations
- **Snapshot Scheduler**: Automated backup tasks
- **Event System**: Task status updates via event bus

This integration allows users to monitor all long-running operations through a unified interface.