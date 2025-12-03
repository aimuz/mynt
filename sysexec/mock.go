package sysexec

import (
	"context"
	"fmt"
	"sync"
)

// MockExecutor is a mock implementation for testing.
type MockExecutor struct {
	mu       sync.Mutex
	commands []Command
	outputs  map[string][]byte
	errors   map[string]error
}

// Command records a command execution.
type Command struct {
	Name string
	Args []string
}

// NewMock creates a new mock command executor.
func NewMock() *MockExecutor {
	return &MockExecutor{
		commands: []Command{},
		outputs:  make(map[string][]byte),
		errors:   make(map[string]error),
	}
}

// SetOutput sets the output for a specific command.
func (m *MockExecutor) SetOutput(name string, output []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputs[name] = output
}

// SetError sets the error for a specific command.
func (m *MockExecutor) SetError(name string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[name] = err
}

// Commands returns all recorded commands.
func (m *MockExecutor) Commands() []Command {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Command{}, m.commands...)
}

// Reset clears all recorded commands and settings.
func (m *MockExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = []Command{}
	m.outputs = make(map[string][]byte)
	m.errors = make(map[string]error)
}

// Run executes a mock command.
func (m *MockExecutor) Run(ctx context.Context, name string, args ...string) error {
	m.mu.Lock()
	m.commands = append(m.commands, Command{Name: name, Args: args})
	err := m.errors[name]
	m.mu.Unlock()
	return err
}

// Output returns mock output for a command.
func (m *MockExecutor) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	m.commands = append(m.commands, Command{Name: name, Args: args})
	output := m.outputs[name]
	err := m.errors[name]
	m.mu.Unlock()

	if err != nil {
		return nil, err
	}
	if output != nil {
		return output, nil
	}
	return fmt.Appendf(nil, "mock output for %s", name), nil
}

// CombinedOutput returns mock combined output for a command.
func (m *MockExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	return m.Output(ctx, name, args...)
}
