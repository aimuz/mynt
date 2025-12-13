package share

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/testutil"
)

// TestIntegration_SambaConfig tests full Samba configuration workflow.
// Requires a real Samba installation to validate config with testparm.
func TestIntegration_SambaConfig(t *testing.T) {
	testutil.RequireIntegration(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "smb.conf")

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	repo := store.NewShareRepo(db)
	mgr := NewManager(repo, configPath)

	tests := []struct {
		name  string
		share store.Share
		want  []string // expected strings in config
	}{
		{
			name: "public share",
			share: store.Share{
				Name:      "public-media",
				Path:      tmpDir,
				Protocol:  "smb",
				Comment:   "Public Media Library",
				ShareType: store.ShareTypePublic,
				ReadOnly:  true,
			},
			want: []string{
				"[public-media]",
				"guest ok = yes",
				"read only = yes",
			},
		},
		{
			name: "restricted share",
			share: store.Share{
				Name:       "finance",
				Path:       tmpDir,
				Protocol:   "smb",
				Comment:    "Finance Dept",
				ShareType:  store.ShareTypeRestricted,
				ValidUsers: "admin,accountant",
			},
			want: []string{
				"[finance]",
				"guest ok = no",
				"valid users = admin,accountant",
			},
		},
		{
			name: "normal share",
			share: store.Share{
				Name:       "projects",
				Path:       tmpDir,
				Protocol:   "smb",
				Comment:    "Project Files",
				ShareType:  store.ShareTypeNormal,
				Browseable: true,
				GuestOK:    false,
			},
			want: []string{
				"[projects]",
				"browseable = yes",
				"guest ok = no",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save share
			s := tt.share
			if err := repo.Save(&s); err != nil {
				t.Fatalf("Save: %v", err)
			}

			// Generate config
			if err := mgr.generateSMBConfig(); err != nil {
				t.Fatalf("generateSMBConfig: %v", err)
			}

			// Read and verify config content
			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			config := string(data)

			for _, want := range tt.want {
				if !strings.Contains(config, want) {
					t.Errorf("config missing %q", want)
				}
			}

			// Validate config with testparm
			if err := mgr.testConfig(); err != nil {
				t.Errorf("testConfig: %v", err)
			}
		})
	}
}

// TestIntegration_ShareLifecycle tests full share CRUD operations.
// Uses generateSMBConfig directly without reloadSamba (requires sudo).
func TestIntegration_ShareLifecycle(t *testing.T) {
	testutil.RequireIntegration(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "smb.conf")

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	repo := store.NewShareRepo(db)
	mgr := NewManager(repo, configPath)

	t.Run("Create", func(t *testing.T) {
		share := &store.Share{
			Name:      "test-lifecycle",
			Path:      tmpDir,
			Protocol:  "smb",
			Comment:   "Lifecycle Test",
			ShareType: store.ShareTypePublic,
		}
		if err := repo.Save(share); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if share.ID == 0 {
			t.Error("ID should be set after save")
		}

		// Generate config
		if err := mgr.generateSMBConfig(); err != nil {
			t.Fatalf("generateSMBConfig: %v", err)
		}

		// Validate with testparm
		if err := mgr.testConfig(); err != nil {
			t.Errorf("testConfig: %v", err)
		}

		// Verify share in config
		data, _ := os.ReadFile(configPath)
		if !strings.Contains(string(data), "[test-lifecycle]") {
			t.Error("share not found in config")
		}
	})

	t.Run("List", func(t *testing.T) {
		shares, err := mgr.ListShares("smb")
		if err != nil {
			t.Fatalf("ListShares: %v", err)
		}
		if len(shares) == 0 {
			t.Error("no shares found")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		shares, _ := mgr.ListShares("smb")
		if len(shares) == 0 {
			t.Skip("no shares to delete")
		}

		id := shares[0].ID
		if err := repo.Delete(id); err != nil {
			t.Fatalf("Delete: %v", err)
		}

		// Regenerate config
		if err := mgr.generateSMBConfig(); err != nil {
			t.Fatalf("generateSMBConfig: %v", err)
		}

		// Validate with testparm
		if err := mgr.testConfig(); err != nil {
			t.Errorf("testConfig: %v", err)
		}
	})
}
