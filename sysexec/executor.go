// Package sysexec provides abstractions for executing external commands.
package sysexec

import "context"

// Executor is an interface for running external commands.
// This abstraction allows for easy mocking in tests and provides a
// consistent API for command execution throughout the application.
type Executor interface {
	// Run executes a command and returns an error if it fails.
	Run(ctx context.Context, name string, args ...string) error

	// Output executes a command and returns its standard output.
	Output(ctx context.Context, name string, args ...string) ([]byte, error)

	// CombinedOutput executes a command and returns its combined stdout and stderr.
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}
