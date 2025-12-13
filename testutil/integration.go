// Package testutil provides test helpers for mynt.
//
// Integration tests require real system resources (ZFS pools, disk devices, etc.)
// and should only run in controlled environments. Use RequireIntegration at the
// beginning of any integration test to skip it during normal test runs.
//
// To run integration tests:
//
//	go test -tags=integration ./...
package testutil

import "testing"

// IntegrationEnabled reports whether integration tests are enabled.
// This is set to true only when the "integration" build tag is active.
var IntegrationEnabled = false

// RequireIntegration skips t unless integration tests are enabled.
// Call this at the start of any test that requires real system resources.
func RequireIntegration(t *testing.T) {
	t.Helper()
	if !IntegrationEnabled {
		t.Skip("skipping integration test (run with -tags=integration)")
	}
}
