// Package testutil provides utilities for testing.
package testutil

import (
	"os"
	"testing"
)

// IsIntegrationTest returns true if integration tests should be run.
// Set INTEGRATION_TESTS=1 environment variable to enable.
func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION_TESTS") == "1"
}

// SkipIfNotIntegration skips the test if not running in integration mode.
func SkipIfNotIntegration(t *testing.T) {
	if !IsIntegrationTest() {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run")
	}
}
