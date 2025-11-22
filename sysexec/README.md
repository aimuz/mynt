# Sysexec Package

This package provides abstractions for executing external system commands.

## Purpose

The `sysexec.Executor` interface provides a clean abstraction over `os/exec`, enabling:

1. **Testability** - Easy mocking of system commands in tests
2. **Consistency** - Uniform API across the codebase
3. **Flexibility** - Swap implementations without changing business logic

## Usage

### Production Code

```go
import "go.aimuz.me/mynt/sysexec"

type DiskManager struct {
    exec sysexec.Executor
}

func NewDiskManager() *DiskManager {
    return &DiskManager{
        exec: sysexec.NewExecutor(), // Real executor
    }
}

func (d *DiskManager) ListDisks(ctx context.Context) ([]Disk, error) {
    output, err := d.exec.Output(ctx, "lsblk", "-J")
    if err != nil {
        return nil, err
    }
    // Parse output...
}
```

### Test Code

```go
import "go.aimuz.me/mynt/sysexec"

func TestDiskManager(t *testing.T) {
    mock := sysexec.NewMock()
    mock.SetOutput("lsblk", []byte(`{"blockdevices": []}`))
    
    mgr := &DiskManager{exec: mock}
    disks, err := mgr.ListDisks(context.Background())
    
    require.NoError(t, err)
    
    // Verify commands executed
    cmds := mock.Commands()
    require.Len(t, cmds, 1)
    require.Equal(t, "lsblk", cmds[0].Name)
}
```

## API

### Executor Interface

```go
type Executor interface {
    Run(ctx context.Context, name string, args ...string) error
    Output(ctx context.Context, name string, args ...string) ([]byte, error)
    CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}
```

### Real Executor

- `sysexec.NewExecutor()` - Creates executor using `os/exec`
- Executes real system commands
- Use in production code

### Mock Executor

- `sysexec.NewMock()` - Creates mock executor
- `SetOutput(name, output)` - Set output for command
- `SetError(name, err)` - Set error for command
- `Commands()` - Get list of executed commands
- `Reset()` - Clear all recorded commands

## Best Practices

1. **Dependency Injection** - Accept `sysexec.Executor` in constructors
2. **Context** - Always pass context for cancellation support
3. **Testing** - Use mock in tests, real executor in production
4. **Error Handling** - Check both error and command exit status

## Migration

To migrate existing code using `exec.Command`:

```go
// Before
cmd := exec.Command("ls", "-la")
output, err := cmd.Output()

// After
type MyService struct {
    exec sysexec.Executor
}

output, err := s.exec.Output(ctx, "ls", "-la")
```
