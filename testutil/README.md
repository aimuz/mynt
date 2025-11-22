# Test Utilities

This package provides utilities for testing Mynt NAS.

## Integration Test Flag

Use the `INTEGRATION_TESTS=1` environment variable to run integration tests that require real system resources.

```bash
# Run only unit tests (skip integration)
go test ./...

# Run all tests including integration
INTEGRATION_TESTS=1 go test ./...

# Or use make target
make test-integration
```

## Usage in Tests

```go
import "go.aimuz.me/mynt/testutil"

func TestWithRealResources(t *testing.T) {
    testutil.SkipIfNotIntegration(t)
    // This test only runs when INTEGRATION_TESTS=1
}
```

## Command Execution Mocking

For mocking command execution, see the `command` package:

```go
import "go.aimuz.me/mynt/command"

// Use mock executor in tests
mock := command.NewMock()
mock.SetOutput("zfs", []byte("pool1"))

// Use real executor in production
exec := command.NewExecutor()
```
