package sysexec

import (
	"context"
	"errors"
	"testing"
)

func TestMockExecutor_Output(t *testing.T) {
	tests := []struct {
		name       string
		cmd        string
		setupMock  func(*MockExecutor)
		wantOutput string
		wantErr    bool
	}{
		{
			name: "default_output",
			cmd:  "test",
			setupMock: func(m *MockExecutor) {
				// No setup - should return default output
			},
			wantOutput: "mock output for test",
			wantErr:    false,
		},
		{
			name: "custom_output",
			cmd:  "ls",
			setupMock: func(m *MockExecutor) {
				m.SetOutput("ls", []byte("file1\nfile2\n"))
			},
			wantOutput: "file1\nfile2\n",
			wantErr:    false,
		},
		{
			name: "error",
			cmd:  "failing",
			setupMock: func(m *MockExecutor) {
				m.SetError("failing", errors.New("command failed"))
			},
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock()
			tt.setupMock(m)

			ctx := context.Background()
			output, err := m.Output(ctx, tt.cmd, "arg1", "arg2")

			if (err != nil) != tt.wantErr {
				t.Errorf("Output() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && string(output) != tt.wantOutput {
				t.Errorf("Output() = %q, want %q", string(output), tt.wantOutput)
			}
		})
	}
}

func TestMockExecutor_Run(t *testing.T) {
	tests := []struct {
		name      string
		cmd       string
		setupMock func(*MockExecutor)
		wantErr   bool
	}{
		{
			name: "success",
			cmd:  "echo",
			setupMock: func(m *MockExecutor) {
				// No error set
			},
			wantErr: false,
		},
		{
			name: "failure",
			cmd:  "false",
			setupMock: func(m *MockExecutor) {
				m.SetError("false", errors.New("exit status 1"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock()
			tt.setupMock(m)

			ctx := context.Background()
			err := m.Run(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockExecutor_CombinedOutput(t *testing.T) {
	m := NewMock()
	m.SetOutput("test", []byte("combined output"))

	ctx := context.Background()
	output, err := m.CombinedOutput(ctx, "test", "arg")

	if err != nil {
		t.Errorf("CombinedOutput() error = %v", err)
	}
	if string(output) != "combined output" {
		t.Errorf("CombinedOutput() = %q, want %q", string(output), "combined output")
	}
}

func TestMockExecutor_Commands(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	// Execute several commands
	m.Run(ctx, "cmd1", "arg1")
	m.Output(ctx, "cmd2", "arg2", "arg3")
	m.CombinedOutput(ctx, "cmd3")

	commands := m.Commands()

	if len(commands) != 3 {
		t.Fatalf("Commands() count = %d, want 3", len(commands))
	}

	tests := []struct {
		idx      int
		wantName string
		wantArgs []string
	}{
		{0, "cmd1", []string{"arg1"}},
		{1, "cmd2", []string{"arg2", "arg3"}},
		{2, "cmd3", []string{}},
	}

	for _, tt := range tests {
		cmd := commands[tt.idx]
		if cmd.Name != tt.wantName {
			t.Errorf("commands[%d].Name = %q, want %q", tt.idx, cmd.Name, tt.wantName)
		}
		if len(cmd.Args) != len(tt.wantArgs) {
			t.Errorf("commands[%d].Args length = %d, want %d", tt.idx, len(cmd.Args), len(tt.wantArgs))
			continue
		}
		for i, arg := range tt.wantArgs {
			if cmd.Args[i] != arg {
				t.Errorf("commands[%d].Args[%d] = %q, want %q", tt.idx, i, cmd.Args[i], arg)
			}
		}
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	// Setup and execute
	m.SetOutput("test", []byte("output"))
	m.SetError("fail", errors.New("error"))
	m.Run(ctx, "cmd1")
	m.Output(ctx, "cmd2")

	if len(m.Commands()) != 2 {
		t.Fatalf("Commands() before reset = %d, want 2", len(m.Commands()))
	}

	// Reset
	m.Reset()

	// Verify reset
	if len(m.Commands()) != 0 {
		t.Errorf("Commands() after reset = %d, want 0", len(m.Commands()))
	}

	// Should return default output after reset
	output, err := m.Output(ctx, "test")
	if err != nil {
		t.Errorf("Output() after reset error = %v", err)
	}
	if string(output) != "mock output for test" {
		t.Errorf("Output() after reset = %q, want default output", string(output))
	}

	// Error should be cleared
	err = m.Run(ctx, "fail")
	if err != nil {
		t.Errorf("Run() after reset should not error, got %v", err)
	}
}

func TestMockExecutor_Concurrent(t *testing.T) {
	m := NewMock()
	m.SetOutput("test", []byte("output"))

	ctx := context.Background()
	done := make(chan bool)

	// Run multiple goroutines
	for i := 0; i < 10; i++ {
		go func(n int) {
			m.Output(ctx, "test")
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	commands := m.Commands()
	if len(commands) != 10 {
		t.Errorf("Commands() count = %d, want 10", len(commands))
	}
}

func TestRealExecutor_Echo(t *testing.T) {
	e := NewExecutor()
	ctx := context.Background()

	// Test with echo command (should be available on all systems)
	output, err := e.Output(ctx, "echo", "hello")
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}

	// Output includes newline from echo
	want := "hello\n"
	if string(output) != want {
		t.Errorf("Output() = %q, want %q", string(output), want)
	}
}

func TestRealExecutor_Run(t *testing.T) {
	e := NewExecutor()
	ctx := context.Background()

	// Test successful command
	err := e.Run(ctx, "true")
	if err != nil {
		t.Errorf("Run(true) error = %v", err)
	}

	// Test failing command
	err = e.Run(ctx, "false")
	if err == nil {
		t.Error("Run(false) expected error, got nil")
	}
}

func TestRealExecutor_CombinedOutput(t *testing.T) {
	e := NewExecutor()
	ctx := context.Background()

	output, err := e.CombinedOutput(ctx, "echo", "test")
	if err != nil {
		t.Fatalf("CombinedOutput() error = %v", err)
	}

	want := "test\n"
	if string(output) != want {
		t.Errorf("CombinedOutput() = %q, want %q", string(output), want)
	}
}

func TestRealExecutor_ContextCancellation(t *testing.T) {
	e := NewExecutor()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := e.Run(ctx, "sleep", "10")
	if err == nil {
		t.Error("Run() with cancelled context should error")
	}
}
