package disk

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// ledctl exit codes are not well documented, but we know:
// 0 = success
// Other codes may indicate various issues

// runLedctl executes ledctl command and handles exit errors gracefully.
func (m *Manager) runLedctl(ctx context.Context, args ...string) error {
	_, err := m.exec.CombinedOutput(ctx, "ledctl", args...)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Log the error but treat non-zero exit as "not supported"
			// since ledctl behavior varies across systems
			code := exitErr.ExitCode()
			return fmt.Errorf("ledctl exited with code %d", code)
		}
		return fmt.Errorf("ledctl: %w", err)
	}
	return nil
}

// Locate turns on the locate LED for a disk.
func (m *Manager) Locate(ctx context.Context, name string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	return m.runLedctl(ctx, "locate=/dev/"+name)
}

// LocateOff turns off the locate LED for a disk.
func (m *Manager) LocateOff(ctx context.Context, name string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	return m.runLedctl(ctx, "locate_off=/dev/"+name)
}
