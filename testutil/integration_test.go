package testutil

import "testing"

func TestIntegrationEnabled(t *testing.T) {
	// Without -tags=integration, IntegrationEnabled should be false.
	// With -tags=integration, IntegrationEnabled should be true.
	// We can't test both in the same binary, but we can verify the current state.
	t.Logf("IntegrationEnabled = %v", IntegrationEnabled)
}

func TestRequireIntegration_Skips(t *testing.T) {
	if IntegrationEnabled {
		t.Skip("test only runs without -tags=integration")
	}

	// Create a sub-test to verify it gets skipped
	t.Run("sub", func(t *testing.T) {
		RequireIntegration(t)
		t.Fatal("should have been skipped")
	})
}

func TestRequireIntegration_Runs(t *testing.T) {
	if !IntegrationEnabled {
		t.Skip("test only runs with -tags=integration")
	}

	// If we get here, RequireIntegration should not skip
	RequireIntegration(t)
	// Success - we weren't skipped
}
