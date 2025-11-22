package sysexec

import (
	"context"
	"os/exec"
)

// RealExecutor executes real system commands using os/exec.
type RealExecutor struct{}

// NewExecutor creates a new real command executor.
func NewExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Run executes a command and returns an error if it fails.
func (e *RealExecutor) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

// Output executes a command and returns its standard output.
func (e *RealExecutor) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// CombinedOutput executes a command and returns its combined stdout and stderr.
func (e *RealExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
