// Package share provides file sharing services.
package share

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/sysexec"
)

// Manager manages file shares (SMB/NFS).
type Manager struct {
	repo       *store.ShareRepo
	exec       sysexec.Executor
	configPath string
	reloadCmd  string
}

// NewManager creates a new share manager.
func NewManager(repo *store.ShareRepo, configPath string) *Manager {
	// Default config path if not specified
	if configPath == "" {
		if runtime.GOOS == "darwin" {
			configPath = "./config/smb.conf" // Development
		} else {
			configPath = "/etc/samba/smb.conf" // Production
		}
	}

	return &Manager{
		repo:       repo,
		exec:       sysexec.NewExecutor(),
		configPath: configPath,
		reloadCmd:  detectSambaReloadCmd(),
	}
}

// CreateShare creates a new SMB share.
func (m *Manager) CreateShare(share *store.Share) error {
	// Validate path exists
	if _, err := os.Stat(share.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", share.Path)
	}

	// Save to database
	if err := m.repo.Save(share); err != nil {
		return fmt.Errorf("failed to save share: %w", err)
	}

	// Regenerate Samba config
	if share.Protocol == "smb" {
		if err := m.generateSMBConfig(); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Reload Samba
		if err := m.reloadSamba(); err != nil {
			return fmt.Errorf("failed to reload samba: %w", err)
		}
	}

	return nil
}

// ListShares returns all shares.
func (m *Manager) ListShares(protocol string) ([]store.Share, error) {
	return m.repo.List(protocol)
}

// DeleteShare removes a share.
func (m *Manager) DeleteShare(id int64) error {
	share, err := m.repo.Get(id)
	if err != nil {
		return err
	}
	if share == nil {
		return fmt.Errorf("share not found")
	}

	// Delete from database
	if err := m.repo.Delete(id); err != nil {
		return err
	}

	// Regenerate config
	if share.Protocol == "smb" {
		if err := m.generateSMBConfig(); err != nil {
			return err
		}
		return m.reloadSamba()
	}

	return nil
}

// generateSMBConfig generates smb.conf from database.
func (m *Manager) generateSMBConfig() error {
	shares, err := m.repo.List("smb")
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	// Global section
	buf.WriteString("[global]\n")
	buf.WriteString("  workgroup = WORKGROUP\n")
	buf.WriteString("  server string = Mynt NAS\n")
	buf.WriteString("  security = user\n")
	buf.WriteString("  map to guest = Bad User\n")
	buf.WriteString("  log file = /var/log/samba/%m.log\n")
	buf.WriteString("  max log size = 50\n\n")

	// Share sections
	for _, share := range shares {
		m.generateShareSection(&buf, share)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write config file
	return os.WriteFile(m.configPath, buf.Bytes(), 0644)
}

// generateShareSection generates Samba config for a single share based on its type
func (m *Manager) generateShareSection(buf *bytes.Buffer, share store.Share) {
	buf.WriteString(fmt.Sprintf("[%s]\n", share.Name))
	buf.WriteString(fmt.Sprintf("  path = %s\n", share.Path))
	buf.WriteString(fmt.Sprintf("  comment = %s\n", share.Comment))

	switch share.ShareType {
	case store.ShareTypePublic:
		// Public share - allow guest access, read-only by default
		buf.WriteString("  browseable = yes\n")
		buf.WriteString("  guest ok = yes\n")
		buf.WriteString(fmt.Sprintf("  read only = %s\n", bStr(share.ReadOnly)))
		buf.WriteString("  create mask = 0644\n")
		buf.WriteString("  directory mask = 0755\n")

	case store.ShareTypeRestricted:
		// Restricted share - only specified users
		buf.WriteString("  browseable = yes\n")
		buf.WriteString("  guest ok = no\n")
		buf.WriteString(fmt.Sprintf("  read only = %s\n", bStr(share.ReadOnly)))
		if share.ValidUsers != "" {
			buf.WriteString(fmt.Sprintf("  valid users = %s\n", share.ValidUsers))
		}
		buf.WriteString("  create mask = 0664\n")
		buf.WriteString("  directory mask = 0775\n")

	default: // ShareTypeNormal
		// Normal share - standard configuration
		buf.WriteString(fmt.Sprintf("  read only = %s\n", bStr(share.ReadOnly)))
		buf.WriteString(fmt.Sprintf("  browseable = %s\n", bStr(share.Browseable)))
		buf.WriteString(fmt.Sprintf("  guest ok = %s\n", bStr(share.GuestOK)))
		if share.ValidUsers != "" {
			buf.WriteString(fmt.Sprintf("  valid users = %s\n", share.ValidUsers))
		}
		buf.WriteString("  create mask = 0664\n")
		buf.WriteString("  directory mask = 0775\n")
	}

	buf.WriteString("\n")
}

// toSambaBoolString converts a boolean to "yes" or "no" string for Samba configuration.
func bStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// reloadSamba reloads the Samba service.
func (m *Manager) reloadSamba() error {
	if m.reloadCmd == "" {
		// Development mode - just validate config
		return m.testConfig()
	}

	ctx := context.Background()
	return m.exec.Run(ctx, "sudo", m.reloadCmd, "smbd", "reload")
}

// testConfig tests the Samba configuration.
func (m *Manager) testConfig() error {
	ctx := context.Background()
	output, err := m.exec.CombinedOutput(ctx, "testparm", "-s", m.configPath)
	if err != nil {
		return fmt.Errorf("config test failed: %s", output)
	}
	return nil
}

// detectSambaReloadCmd detects the correct command to reload Samba.
func detectSambaReloadCmd() string {
	ctx := context.Background()
	exec := sysexec.NewExecutor()

	// Try systemctl first (modern Linux)
	_, err := exec.Output(ctx, "which", "systemctl")
	if err == nil {
		return "systemctl"
	}

	// Try service (older Linux)
	_, err = exec.Output(ctx, "which", "service")
	if err == nil {
		return "service"
	}

	// macOS or development - no reload
	return ""
}
