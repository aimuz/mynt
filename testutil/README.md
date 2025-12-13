# Package testutil

Test helpers for mynt.

## Integration Tests

Integration tests require real system resources (ZFS pools, disk devices, etc.)
and are skipped by default.

```bash
# Run unit tests only (default)
go test ./...

# Run all tests including integration
go test -tags=integration ./...
```

## Usage

```go
func TestZFSPoolCreate(t *testing.T) {
    testutil.RequireIntegration(t)
    // Test runs only with -tags=integration
}
```

## Design

The package uses a simple pattern:

- `integration.go` — defines `IntegrationEnabled` (default false) and `RequireIntegration`
- `integration_on.go` — build-tagged file that sets `IntegrationEnabled = true` via init

This follows the Go standard library pattern (e.g., `race.go`/`race0.go`).
