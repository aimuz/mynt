package disk

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"go.aimuz.me/mynt/sysexec"
)

func TestManager_Locate(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Locate is a no-op on darwin")
	}

	tests := []struct {
		name      string
		diskName  string
		mockErr   error
		wantErr   bool
	}{
		{
			name:     "success",
			diskName: "sda",
			mockErr:  nil,
			wantErr:  false,
		},
		{
			name:     "command_failure",
			diskName: "sdb",
			mockErr:  errors.New("ledctl error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := sysexec.NewMock()
			if tt.mockErr != nil {
				mock.SetError("ledctl", tt.mockErr)
			} else {
				mock.SetOutput("ledctl", []byte(""))
			}

			m := &Manager{exec: mock}

			ctx := context.Background()
			err := m.Locate(ctx, tt.diskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Locate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify command was called correctly
				commands := mock.Commands()
				if len(commands) != 1 {
					t.Fatalf("expected 1 command, got %d", len(commands))
				}
				if commands[0].Name != "ledctl" {
					t.Errorf("command name = %q, want %q", commands[0].Name, "ledctl")
				}
				expectedArg := "locate=/dev/" + tt.diskName
				if len(commands[0].Args) != 1 || commands[0].Args[0] != expectedArg {
					t.Errorf("command args = %v, want [%s]", commands[0].Args, expectedArg)
				}
			}
		})
	}
}

func TestManager_LocateOff(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("LocateOff is a no-op on darwin")
	}

	tests := []struct {
		name      string
		diskName  string
		mockErr   error
		wantErr   bool
	}{
		{
			name:     "success",
			diskName: "sda",
			mockErr:  nil,
			wantErr:  false,
		},
		{
			name:     "command_failure",
			diskName: "sdb",
			mockErr:  errors.New("ledctl error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := sysexec.NewMock()
			if tt.mockErr != nil {
				mock.SetError("ledctl", tt.mockErr)
			} else {
				mock.SetOutput("ledctl", []byte(""))
			}

			m := &Manager{exec: mock}

			ctx := context.Background()
			err := m.LocateOff(ctx, tt.diskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("LocateOff() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify command was called correctly
				commands := mock.Commands()
				if len(commands) != 1 {
					t.Fatalf("expected 1 command, got %d", len(commands))
				}
				if commands[0].Name != "ledctl" {
					t.Errorf("command name = %q, want %q", commands[0].Name, "ledctl")
				}
				expectedArg := "locate_off=/dev/" + tt.diskName
				if len(commands[0].Args) != 1 || commands[0].Args[0] != expectedArg {
					t.Errorf("command args = %v, want [%s]", commands[0].Args, expectedArg)
				}
			}
		})
	}
}

func TestManager_Locate_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("This test is for darwin only")
	}

	m := NewManager()
	ctx := context.Background()

	// Should not error on darwin
	err := m.Locate(ctx, "disk0")
	if err != nil {
		t.Errorf("Locate() on darwin should not error, got %v", err)
	}

	err = m.LocateOff(ctx, "disk0")
	if err != nil {
		t.Errorf("LocateOff() on darwin should not error, got %v", err)
	}
}
