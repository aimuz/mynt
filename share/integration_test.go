package share

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/testutil"
)

// TestIntegration_SambaConfig validates generated config with real testparm.
// This is the unique value of integration tests - unit tests can't run testparm.
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

	// Create multiple share types to validate testparm handles them all
	shares := []store.Share{
		{Name: "public", Path: tmpDir, Protocol: "smb", ShareType: store.ShareTypePublic},
		{Name: "restricted", Path: tmpDir, Protocol: "smb", ShareType: store.ShareTypeRestricted, ValidUsers: "admin"},
		{Name: "normal", Path: tmpDir, Protocol: "smb", ShareType: store.ShareTypeNormal},
	}

	for i := range shares {
		if err := repo.Save(&shares[i]); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	if err := mgr.generateSMBConfig(); err != nil {
		t.Fatalf("generateSMBConfig: %v", err)
	}

	// The real test: validate with testparm
	if err := mgr.testConfig(); err != nil {
		t.Fatalf("testparm validation failed: %v", err)
	}
}

// TestIntegration_ShareLifecycle tests share CRUD with real testparm validation.
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

	var shareID int64

	t.Run("Create", func(t *testing.T) {
		share := &store.Share{
			Name:      "lifecycle-test",
			Path:      tmpDir,
			Protocol:  "smb",
			ShareType: store.ShareTypePublic,
		}
		if err := repo.Save(share); err != nil {
			t.Fatalf("Save: %v", err)
		}
		shareID = share.ID

		if err := mgr.generateSMBConfig(); err != nil {
			t.Fatalf("generateSMBConfig: %v", err)
		}
		if err := mgr.testConfig(); err != nil {
			t.Fatalf("testparm: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		shares, err := mgr.ListShares("smb")
		if err != nil {
			t.Fatalf("ListShares: %v", err)
		}
		if len(shares) != 1 {
			t.Errorf("len(shares) = %d, want 1", len(shares))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(shareID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if err := mgr.generateSMBConfig(); err != nil {
			t.Fatalf("generateSMBConfig: %v", err)
		}
		if err := mgr.testConfig(); err != nil {
			t.Fatalf("testparm: %v", err)
		}

		// Verify share removed (this is lifecycle-specific, not content verification)
		data, _ := os.ReadFile(configPath)
		if strings.Contains(string(data), "[lifecycle-test]") {
			t.Error("deleted share still in config")
		}
	})
}
