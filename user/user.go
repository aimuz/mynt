// Package user manages system and virtual user accounts.
package user

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"strings"

	"go.aimuz.me/mynt/logger"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/sysexec"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrMissingRequired    = errors.New("username and password are required")
)

// CreateRequest is a request to create a user.
type CreateRequest struct {
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	FullName    string            `json:"full_name"`
	Email       string            `json:"email"`
	AccountType store.AccountType `json:"account_type"`
	IsAdmin     bool              `json:"is_admin"`
}

// Manager manages user accounts.
type Manager struct {
	repo *store.UserRepo
	exec sysexec.Executor
}

// NewManager returns a new Manager.
func NewManager(repo *store.UserRepo) *Manager {
	return &Manager{
		repo: repo,
		exec: sysexec.NewExecutor(),
	}
}

// SetExecutor sets the command executor for testing.
func (m *Manager) SetExecutor(exec sysexec.Executor) {
	m.exec = exec
}

// Create creates a new user.
func (m *Manager) Create(req CreateRequest) (*store.User, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrMissingRequired
	}

	existing, err := m.repo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	if req.AccountType == "" {
		req.AccountType = store.AccountVirtual
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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

	if req.AccountType == store.AccountSystem {
		if err := m.createSystemUser(user, req.Password); err != nil {
			return nil, err
		}
	}

	if err := m.syncSambaUser(req.Username, req.Password); err != nil {
		logger.Warn("failed to sync samba user",
			"username", req.Username,
			"error", err)
	}

	if err := m.repo.Save(user); err != nil {
		if req.AccountType == store.AccountSystem {
			m.deleteSystemUser(req.Username)
		}
		return nil, err
	}

	return user, nil
}

// List returns all users.
func (m *Manager) List() ([]store.User, error) {
	return m.repo.List()
}

// Get returns a user by username.
func (m *Manager) Get(username string) (*store.User, error) {
	return m.repo.GetByUsername(username)
}

// Delete deletes a user.
func (m *Manager) Delete(username string) error {
	user, err := m.repo.GetByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := m.repo.Delete(user.ID); err != nil {
		return err
	}

	if user.AccountType == store.AccountSystem {
		m.deleteSystemUser(username)
	}

	m.deleteSambaUser(username)

	return nil
}

// VerifyPassword checks credentials and returns the user if valid.
func (m *Manager) VerifyPassword(username, password string) (*store.User, error) {
	user, err := m.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	m.repo.UpdateLastLogin(user.ID)

	return user, nil
}

func (m *Manager) createSystemUser(user *store.User, password string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}

	ctx := context.Background()
	err := m.exec.Run(ctx, "useradd",
		"-m",
		"-s", "/bin/bash",
		"-c", user.FullName,
		user.Username,
	)
	if err != nil {
		return err
	}

	input := user.Username + ":" + password
	err = m.exec.Run(ctx, "sh", "-c", "echo '"+input+"' | chpasswd")
	if err != nil {
		m.deleteSystemUser(user.Username)
		return err
	}

	user.UID, user.GID, _ = m.getUserIDs(ctx, user.Username)
	user.HomeDir = "/home/" + user.Username
	user.Shell = "/bin/bash"

	return nil
}

func (m *Manager) deleteSystemUser(username string) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	return m.exec.Run(context.Background(), "userdel", "-r", username)
}

func (m *Manager) syncSambaUser(username, password string) error {
	input := password + "\\n" + password + "\\n"
	cmd := "echo -e '" + input + "' | smbpasswd -a -s " + username
	return m.exec.Run(context.Background(), "sh", "-c", cmd)
}

func (m *Manager) deleteSambaUser(username string) error {
	return m.exec.Run(context.Background(), "smbpasswd", "-x", username)
}

func (m *Manager) getUserIDs(ctx context.Context, username string) (*int, *int, error) {
	out, err := m.exec.Output(ctx, "id", "-u", username)
	if err != nil {
		return nil, nil, err
	}
	uid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return nil, nil, err
	}

	out, err = m.exec.Output(ctx, "id", "-g", username)
	if err != nil {
		return nil, nil, err
	}
	gid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return nil, nil, err
	}

	return &uid, &gid, nil
}
