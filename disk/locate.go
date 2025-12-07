package disk

import (
	"context"
	"os/exec"
	"runtime"
)

// Locate turns on the locate LED for a disk.
func (m *Manager) Locate(ctx context.Context, name string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	_, err := m.exec.CombinedOutput(ctx, "ledctl", "locate=/dev/"+name)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() <= 3 {
			return nil
		}
		return err
	}
	return nil
}

// LocateOff turns off the locate LED for a disk.
func (m *Manager) LocateOff(ctx context.Context, name string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	_, err := m.exec.CombinedOutput(ctx, "ledctl", "locate_off=/dev/"+name)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() <= 3 {
			return nil
		}
		return err
	}
	return nil
}
