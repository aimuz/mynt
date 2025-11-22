// Package user provides user management functionality.
package user

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"go.aimuz.me/mynt/logger"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/sysexec"
	"golang.org/x/crypto/bcrypt"
)

// CreateRequest represents a request to create a user.
type CreateRequest struct {
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	FullName    string            `json:"full_name"`
	Email       string            `json:"email"`
	AccountType store.AccountType `json:"account_type"` // "virtual" or "system"
	IsAdmin     bool              `json:"is_admin"`
}

// Manager handles user operations.
type Manager struct {
	repo *store.UserRepo
	exec sysexec.Executor
}

// NewManager creates a new user manager.
func NewManager(repo *store.UserRepo) *Manager {
	return &Manager{
		repo: repo,
		exec: sysexec.NewExecutor(),
	}
}

// SetExecutor sets a custom command executor (for testing).
func (m *Manager) setExecutor(exec sysexec.Executor) {
	m.exec = exec
}

// Create creates a new user.
func (m *Manager) Create(req CreateRequest) (*store.User, error) {
	// Validate
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	// Check if user exists
	existing, err := m.repo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Default to virtual account
	if req.AccountType == "" {
		req.AccountType = store.AccountVirtual
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &store.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Email:        req.Email,
		AccountType:  req.AccountType,
		IsAdmin:      req.IsAdmin,
		IsActive:     true,
	}

	ctx := context.Background()

	// Create system user if needed
	if req.AccountType == store.AccountSystem {
		if err := m.createSystemUser(ctx, user, req.Password); err != nil {
			return nil, fmt.Errorf("failed to create system user: %w", err)
		}
	}

	// Always sync to Samba (both virtual and system accounts can access SMB)
	if err := m.syncSambaUser(ctx, req.Username, req.Password); err != nil {
		// Log error but don't fail - Samba might not be installed or configured
		logger.Warn("failed to sync samba user",
			"username", req.Username,
			"error", err)
	}

	// Save to database
	if err := m.repo.Save(user); err != nil {
		// Cleanup system user if DB save fails
		if req.AccountType == store.AccountSystem {
			m.deleteSystemUser(ctx, req.Username)
		}
		return nil, err
	}

	return user, nil
}

// List returns all users.
func (m *Manager) List() ([]store.User, error) {
	return m.repo.List()
}

// Get retrieves a user by username.
func (m *Manager) Get(username string) (*store.User, error) {
	return m.repo.GetByUsername(username)
}

// Delete removes a user.
func (m *Manager) Delete(username string) error {
	user, err := m.repo.GetByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	ctx := context.Background()

	// Delete from database first
	if err := m.repo.Delete(user.ID); err != nil {
		return err
	}

	// Delete system user if it exists
	if user.AccountType == store.AccountSystem {
		m.deleteSystemUser(ctx, username)
	}

	// Delete Samba user
	m.deleteSambaUser(ctx, username)

	return nil
}

// VerifyPassword verifies a user's password.
func (m *Manager) VerifyPassword(username, password string) (*store.User, error) {
	user, err := m.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	m.repo.UpdateLastLogin(user.ID)

	return user, nil
}

// createSystemUser creates a Linux system user.
func (m *Manager) createSystemUser(ctx context.Context, user *store.User, password string) error {
	if runtime.GOOS == "darwin" {
		// macOS - skip system user creation in development
		return nil
	}

	// Create user with useradd
	err := m.exec.Run(ctx, "useradd",
		"-m",              // Create home directory
		"-s", "/bin/bash", // Set shell
		"-c", user.FullName, // Full name (comment)
		user.Username,
	)
	if err != nil {
		return fmt.Errorf("useradd failed: %w", err)
	}

	// Set password using chpasswd (reads from stdin)
	passwordInput := fmt.Sprintf("%s:%s", user.Username, password)
	err = m.exec.Run(ctx, "sh", "-c", fmt.Sprintf("echo '%s' | chpasswd", passwordInput))
	if err != nil {
		m.deleteSystemUser(ctx, user.Username) // Cleanup
		return fmt.Errorf("chpasswd failed: %w", err)
	}

	// Get UID/GID
	user.UID, user.GID, _ = m.getUserIDs(ctx, user.Username)
	user.HomeDir = "/home/" + user.Username
	user.Shell = "/bin/bash"

	return nil
}

// deleteSystemUser removes a Linux system user.
func (m *Manager) deleteSystemUser(ctx context.Context, username string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}

	return m.exec.Run(ctx, "userdel", "-r", username) // -r removes home directory
}

// syncSambaUser adds/updates a Samba user.
func (m *Manager) syncSambaUser(ctx context.Context, username, password string) error {
	// Use smbpasswd in non-interactive mode
	passwordInput := password + "\\n" + password + "\\n"
	err := m.exec.Run(ctx, "sh", "-c", fmt.Sprintf("echo -e '%s' | smbpasswd -a -s %s", passwordInput, username))
	return err
}

// deleteSambaUser removes a Samba user.
func (m *Manager) deleteSambaUser(ctx context.Context, username string) error {
	return m.exec.Run(ctx, "smbpasswd", "-x", username)
}

// getUserIDs gets UID and GID for a username.
func (m *Manager) getUserIDs(ctx context.Context, username string) (*int, *int, error) {
	uidOut, err := m.exec.Output(ctx, "id", "-u", username)
	if err != nil {
		return nil, nil, err
	}

	uid, err := strconv.Atoi(strings.TrimSpace(string(uidOut)))
	if err != nil {
		return nil, nil, err
	}

	gidOut, err := m.exec.Output(ctx, "id", "-g", username)
	if err != nil {
		return nil, nil, err
	}

	gid, err := strconv.Atoi(strings.TrimSpace(string(gidOut)))
	if err != nil {
		return nil, nil, err
	}

	return &uid, &gid, nil
}
