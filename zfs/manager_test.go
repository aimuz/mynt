package zfs

import (
	"context"
	"fmt"
	"testing"

	"go.aimuz.me/mynt/sysexec"
)

func TestListPools(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput []byte
		mockError  error
		wantErr    bool
		wantCount  int
		checkPool  func(t *testing.T, pool Pool)
	}{
		{
			name: "valid_pool_list",
			mockOutput: []byte("tank	12345678	1099511627776	524288000000	575223627776	5%	ONLINE	-\n" +
				"backup	87654321	2199023255552	1048576000000	1150447255552	10%	DEGRADED	/mnt/alt\n"),
			wantErr:   false,
			wantCount: 2,
			checkPool: func(t *testing.T, pool Pool) {
				if pool.Name == "tank" {
					if pool.GUID != "12345678" {
						t.Errorf("expected GUID to be '12345678', got %s", pool.GUID)
					}
					if pool.Size != 1099511627776 {
						t.Errorf("expected Size to be 1099511627776, got %d", pool.Size)
					}
					if pool.Health != PoolOnline {
						t.Errorf("expected Health to be ONLINE, got %s", pool.Health)
					}
					if pool.Frag != 5 {
						t.Errorf("expected Frag to be 5, got %d", pool.Frag)
					}
				}
			},
		},
		{
			name:       "empty_output",
			mockOutput: []byte(""),
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:      "command_error",
			mockError: fmt.Errorf("zpool command failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			if tt.mockError != nil {
				mockExec.SetError("zpool", tt.mockError)
			} else {
				mockExec.SetOutput("zpool", tt.mockOutput)
			}
			m := &Manager{exec: mockExec}

			pools, err := m.ListPools(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(pools) != tt.wantCount {
					t.Errorf("expected %d pools, got %d", tt.wantCount, len(pools))
				}

				if tt.checkPool != nil && len(pools) > 0 {
					tt.checkPool(t, pools[0])
				}

				// Verify command arguments
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				if cmds[0].Name != "zpool" {
					t.Errorf("expected command name to be 'zpool', got %s", cmds[0].Name)
				}
				expectedArgs := []string{"list", "-H", "-p", "-o", "name,guid,size,alloc,free,frag,health,altroot"}
				if !stringSliceEqual(cmds[0].Args, expectedArgs) {
					t.Errorf("expected args %v, got %v", expectedArgs, cmds[0].Args)
				}
			}
		})
	}
}

func TestListDatasets(t *testing.T) {
	tests := []struct {
		name         string
		mockOutput   []byte
		mockError    error
		wantErr      bool
		wantCount    int
		checkDataset func(t *testing.T, dataset Dataset)
	}{
		{
			name: "valid_dataset_list",
			mockOutput: []byte("pool/dataset1	filesystem	1048576000	2097152000	524288000	/pool/dataset1	lz4	off	off\n" +
				"pool/dataset2	volume	2097152000	4194304000	1048576000	-	gzip	on	on\n"),
			wantErr:   false,
			wantCount: 2,
			checkDataset: func(t *testing.T, dataset Dataset) {
				if dataset.Name == "pool/dataset1" {
					if dataset.Type != DatasetFilesystem {
						t.Errorf("expected Type to be filesystem, got %s", dataset.Type)
					}
					if dataset.Used != 1048576000 {
						t.Errorf("expected Used to be 1048576000, got %d", dataset.Used)
					}
					if dataset.Compression != "lz4" {
						t.Errorf("expected Compression to be lz4, got %s", dataset.Compression)
					}
					if dataset.Mountpoint != "/pool/dataset1" {
						t.Errorf("expected Mountpoint to be /pool/dataset1, got %s", dataset.Mountpoint)
					}
				}
			},
		},
		{
			name:       "empty_output",
			mockOutput: []byte(""),
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:      "command_error",
			mockError: fmt.Errorf("zfs command failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			if tt.mockError != nil {
				mockExec.SetError("zfs", tt.mockError)
			} else {
				mockExec.SetOutput("zfs", tt.mockOutput)
			}
			m := &Manager{exec: mockExec}

			datasets, err := m.ListDatasets(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListDatasets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(datasets) != tt.wantCount {
					t.Errorf("expected %d datasets, got %d", tt.wantCount, len(datasets))
				}

				if tt.checkDataset != nil && len(datasets) > 0 {
					tt.checkDataset(t, datasets[0])
				}

				// Verify command arguments
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				expectedArgs := []string{"list", "-H", "-p", "-o", "name,type,used,avail,refer,mountpoint,compression,encryption,dedup"}
				if !stringSliceEqual(cmds[0].Args, expectedArgs) {
					t.Errorf("expected args %v, got %v", expectedArgs, cmds[0].Args)
				}
			}
		})
	}
}

func TestCreatePool(t *testing.T) {
	tests := []struct {
		name      string
		req       CreatePoolRequest
		wantErr   bool
		checkArgs func(t *testing.T, args []string)
	}{
		{
			name: "simple_pool_no_type",
			req: CreatePoolRequest{
				Name:    "tank",
				Devices: []string{"/dev/sda", "/dev/sdb"},
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				expected := []string{"create", "-f", "tank", "/dev/sda", "/dev/sdb"}
				if !stringSliceEqual(args, expected) {
					t.Errorf("expected args %v, got %v", expected, args)
				}
			},
		},
		{
			name: "mirror_pool",
			req: CreatePoolRequest{
				Name:    "tank",
				Type:    "mirror",
				Devices: []string{"/dev/sda", "/dev/sdb"},
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				expected := []string{"create", "-f", "tank", "mirror", "/dev/sda", "/dev/sdb"}
				if !stringSliceEqual(args, expected) {
					t.Errorf("expected args %v, got %v", expected, args)
				}
			},
		},
		{
			name: "raidz_pool",
			req: CreatePoolRequest{
				Name:    "backup",
				Type:    "raidz",
				Devices: []string{"/dev/sdc", "/dev/sdd", "/dev/sde"},
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				if args[0] != "create" || args[1] != "-f" || args[2] != "backup" || args[3] != "raidz" {
					t.Errorf("unexpected args: %v", args)
				}
				if len(args) != 7 {
					t.Errorf("expected 7 args, got %d", len(args))
				}
			},
		},
		{
			name: "single_device",
			req: CreatePoolRequest{
				Name:    "test",
				Devices: []string{"/dev/sdf"},
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				expected := []string{"create", "-f", "test", "/dev/sdf"}
				if !stringSliceEqual(args, expected) {
					t.Errorf("expected args %v, got %v", expected, args)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			m := &Manager{exec: mockExec}

			err := m.CreatePool(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				if cmds[0].Name != "zpool" {
					t.Errorf("expected command name to be 'zpool', got %s", cmds[0].Name)
				}
				if tt.checkArgs != nil {
					tt.checkArgs(t, cmds[0].Args)
				}
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.exec == nil {
		t.Error("Manager.exec should not be nil")
	}
}

// Helper function to compare string slices
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
