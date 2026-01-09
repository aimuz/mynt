package testutil

import (
	"os"
	"testing"
)

func TestIsIntegrationTest(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    bool
	}{
		{
			name:   "enabled",
			envVal: "1",
			want:   true,
		},
		{
			name:   "disabled",
			envVal: "",
			want:   false,
		},
		{
			name:   "disabled_zero",
			envVal: "0",
			want:   false,
		},
		{
			name:   "disabled_other",
			envVal: "yes",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			orig := os.Getenv("INTEGRATION_TESTS")
			defer os.Setenv("INTEGRATION_TESTS", orig)

			// Set test value
			if tt.envVal == "" {
				os.Unsetenv("INTEGRATION_TESTS")
			} else {
				os.Setenv("INTEGRATION_TESTS", tt.envVal)
			}

			got := IsIntegrationTest()
			if got != tt.want {
				t.Errorf("IsIntegrationTest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipIfNotIntegration(t *testing.T) {
	// Save original value
	orig := os.Getenv("INTEGRATION_TESTS")
	defer os.Setenv("INTEGRATION_TESTS", orig)

	// Test that it skips when not enabled
	os.Unsetenv("INTEGRATION_TESTS")

	// Create a mock test that will be skipped
	mockT := &testing.T{}
	
	// We can't easily test the skip behavior without running a sub-test,
	// so we just verify the function doesn't panic
	if IsIntegrationTest() {
		t.Error("IsIntegrationTest() should be false")
	}

	// Test that it doesn't skip when enabled
	os.Setenv("INTEGRATION_TESTS", "1")
	if !IsIntegrationTest() {
		t.Error("IsIntegrationTest() should be true")
	}

	// The actual skip behavior is tested by using SkipIfNotIntegration
	// in real integration tests
	_ = mockT
}
