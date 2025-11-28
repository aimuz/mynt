package zfs

import (
	"context"
	"strings"
	"testing"

	"go.aimuz.me/mynt/sysexec"
)

func TestCreateDataset(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateDatasetRequest
		wantErr     bool
		errContains string
		checkArgs   func(t *testing.T, args []string)
	}{
		{
			name: "valid_filesystem_with_properties",
			req: CreateDatasetRequest{
				Name: "pool/dataset1",
				Type: "filesystem",
				Properties: map[string]string{
					"compression": "lz4",
					"atime":       "off",
				},
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				if args[0] != "create" {
					t.Errorf("expected first arg to be 'create', got %s", args[0])
				}
				// Check that properties are included
				propsFound := 0
				for i, arg := range args {
					if arg == "-o" {
						propsFound++
						propValue := args[i+1]
						if !strings.HasPrefix(propValue, "compression=") && !strings.HasPrefix(propValue, "atime=") {
							t.Errorf("unexpected property: %s", propValue)
						}
					}
				}
				if propsFound != 2 {
					t.Errorf("expected 2 properties, found %d", propsFound)
				}
				// Check dataset name is last
				if args[len(args)-1] != "pool/dataset1" {
					t.Errorf("expected last arg to be dataset name, got %s", args[len(args)-1])
				}
			},
		},
		{
			name: "valid_volume_with_size",
			req: CreateDatasetRequest{
				Name: "pool/volume1",
				Type: "volume",
				Size: 10 * 1024 * 1024 * 1024, // 10GB
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				hasV := false
				for i, arg := range args {
					if arg == "-V" {
						hasV = true
						sizeStr := args[i+1]
						if sizeStr != "10G" {
							t.Errorf("expected size to be 10G, got %s", sizeStr)
						}
					}
				}
				if !hasV {
					t.Error("expected -V flag for volume")
				}
			},
		},
		{
			name: "default_type_to_filesystem",
			req: CreateDatasetRequest{
				Name: "pool/dataset2",
			},
			wantErr: false,
			checkArgs: func(t *testing.T, args []string) {
				// No -V flag should be present
				for _, arg := range args {
					if arg == "-V" {
						t.Error("should not have -V flag for filesystem")
					}
				}
			},
		},
		{
			name: "missing_name",
			req: CreateDatasetRequest{
				Type: "filesystem",
			},
			wantErr:     true,
			errContains: "dataset name is required",
		},
		{
			name: "volume_without_size",
			req: CreateDatasetRequest{
				Name: "pool/volume2",
				Type: "volume",
			},
			wantErr:     true,
			errContains: "size is required for volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			m := &Manager{exec: mockExec}

			err := m.CreateDataset(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// Verify command was called correctly
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				if cmds[0].Name != "zfs" {
					t.Errorf("expected command name to be 'zfs', got %s", cmds[0].Name)
				}
				if tt.checkArgs != nil {
					tt.checkArgs(t, cmds[0].Args)
				}
			}
		})
	}
}

func TestGetDataset(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		mockOutput  []byte
		wantErr     bool
		errContains string
	}{
		{
			name:        "found_dataset",
			datasetName: "pool/dataset1",
			mockOutput: []byte("pool/dataset1	filesystem	1024	2048	512	/pool/dataset1	lz4	off	off\n" +
				"pool/dataset2	filesystem	2048	4096	1024	/pool/dataset2	gzip	on	on\n"),
			wantErr: false,
		},
		{
			name:        "dataset_not_found",
			datasetName: "pool/nonexistent",
			mockOutput: []byte("pool/dataset1	filesystem	1024	2048	512	/pool/dataset1	lz4	off	off\n" +
				"pool/dataset2	filesystem	2048	4096	1024	/pool/dataset2	gzip	on	on\n"),
			wantErr:     true,
			errContains: "dataset not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			mockExec.SetOutput("zfs", tt.mockOutput)
			m := &Manager{exec: mockExec}

			dataset, err := m.GetDataset(context.Background(), tt.datasetName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if dataset == nil {
					t.Fatal("expected dataset to be non-nil")
				}
				if dataset.Name != tt.datasetName {
					t.Errorf("expected dataset name to be %s, got %s", tt.datasetName, dataset.Name)
				}
			}
		})
	}
}

func TestDestroyDataset(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		wantErr     bool
		errContains string
		checkArgs   func(t *testing.T, args []string)
	}{
		{
			name:        "valid_destroy",
			datasetName: "pool/dataset1",
			wantErr:     false,
			checkArgs: func(t *testing.T, args []string) {
				expected := []string{"destroy", "-r", "pool/dataset1"}
				if len(args) != len(expected) {
					t.Fatalf("expected %d args, got %d", len(expected), len(args))
				}
				for i, arg := range expected {
					if args[i] != arg {
						t.Errorf("arg[%d]: expected %s, got %s", i, arg, args[i])
					}
				}
			},
		},
		{
			name:        "missing_name",
			datasetName: "",
			wantErr:     true,
			errContains: "dataset name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			m := &Manager{exec: mockExec}

			err := m.DestroyDataset(context.Background(), tt.datasetName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DestroyDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				if tt.checkArgs != nil {
					tt.checkArgs(t, cmds[0].Args)
				}
			}
		})
	}
}

func TestSetProperty(t *testing.T) {
	tests := []struct {
		name        string
		datasetName string
		key         string
		value       string
		wantErr     bool
		errContains string
		checkArgs   func(t *testing.T, args []string)
	}{
		{
			name:        "valid_set_property",
			datasetName: "pool/dataset1",
			key:         "compression",
			value:       "lz4",
			wantErr:     false,
			checkArgs: func(t *testing.T, args []string) {
				expected := []string{"set", "compression=lz4", "pool/dataset1"}
				if len(args) != len(expected) {
					t.Fatalf("expected %d args, got %d", len(expected), len(args))
				}
				for i, arg := range expected {
					if args[i] != arg {
						t.Errorf("arg[%d]: expected %s, got %s", i, arg, args[i])
					}
				}
			},
		},
		{
			name:        "missing_dataset_name",
			datasetName: "",
			key:         "compression",
			value:       "lz4",
			wantErr:     true,
			errContains: "dataset name and property key are required",
		},
		{
			name:        "missing_property_key",
			datasetName: "pool/dataset1",
			key:         "",
			value:       "lz4",
			wantErr:     true,
			errContains: "dataset name and property key are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := sysexec.NewMock()
			m := &Manager{exec: mockExec}

			err := m.SetProperty(context.Background(), tt.datasetName, tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				cmds := mockExec.Commands()
				if len(cmds) != 1 {
					t.Fatalf("expected 1 command, got %d", len(cmds))
				}
				if tt.checkArgs != nil {
					tt.checkArgs(t, cmds[0].Args)
				}
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 512,
			want:  "512",
		},
		{
			name:  "kilobytes",
			bytes: 1024,
			want:  "1K",
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024,
			want:  "1M",
		},
		{
			name:  "gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1G",
		},
		{
			name:  "terabytes",
			bytes: 1024 * 1024 * 1024 * 1024,
			want:  "1T",
		},
		{
			name:  "10_gigabytes",
			bytes: 10 * 1024 * 1024 * 1024,
			want:  "10G",
		},
		{
			name:  "rounding_down",
			bytes: 1536, // 1.5K
			want:  "1K",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, got, tt.want)
			}
		})
	}
}
