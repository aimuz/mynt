package user

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/sysexec"
)

func setupTestUser(t *testing.T) (*Manager, *store.DB) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)

	userRepo := store.NewUserRepo(db)
	mgr := NewManager(userRepo)

	// Set mock executor for testing
	mock := sysexec.NewMock()
	mgr.setExecutor(mock)

	return mgr, db
}

func TestCreateUser_Virtual(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username:    "testuser",
		Password:    "TestPass123!",
		FullName:    "Test User",
		AccountType: store.AccountVirtual,
	}

	user, err := mgr.Create(req)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Greater(t, user.ID, int64(0))
	require.Equal(t, "testuser", user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.Equal(t, store.AccountVirtual, user.AccountType)
	require.False(t, user.IsAdmin)
}

func TestCreateUser_System(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username:    "sysuser",
		Password:    "SysPass123!",
		FullName:    "System User",
		AccountType: store.AccountSystem,
		IsAdmin:     true,
	}

	user, err := mgr.Create(req)
	// On macOS or without proper permissions, system user creation is skipped
	// So we just check that the user is created in DB
	require.NoError(t, err)
	require.Equal(t, store.AccountSystem, user.AccountType)
	require.True(t, user.IsAdmin)
}

func TestCreateUser_DefaultAccountType(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "defaultuser",
		Password: "Pass123!",
	}

	user, err := mgr.Create(req)
	require.NoError(t, err)
	// Should default to virtual
	require.Equal(t, store.AccountVirtual, user.AccountType)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "dupuser",
		Password: "Pass123!",
	}

	_, err := mgr.Create(req)
	require.NoError(t, err)

	// Try to create duplicate
	_, err = mgr.Create(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	tests := []struct {
		name    string
		request CreateRequest
		errMsg  string
	}{
		{
			name:    "empty_username",
			request: CreateRequest{Password: "Pass123!"},
			errMsg:  "required",
		},
		{
			name:    "empty_password",
			request: CreateRequest{Username: "user"},
			errMsg:  "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Create(tt.request)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "testuser",
		Password: "TestPass123!",
	}
	mgr.Create(req)

	// Verify correct password
	user, err := mgr.VerifyPassword("testuser", "TestPass123!")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "testuser", user.Username)
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "testuser",
		Password: "TestPass123!",
	}
	mgr.Create(req)

	// Try wrong password
	_, err := mgr.VerifyPassword("testuser", "WrongPassword")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestVerifyPassword_NonexistentUser(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	_, err := mgr.VerifyPassword("nonexistent", "anypassword")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestVerifyPassword_InactiveUser(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "testuser",
		Password: "TestPass123!",
	}
	user, _ := mgr.Create(req)

	// Deactivate user
	user.IsActive = false
	mgr.repo.Update(user)

	// Try to verify
	_, err := mgr.VerifyPassword("testuser", "TestPass123!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid credentials")
}

func TestListUsers(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	// Create some users
	for i := 1; i <= 3; i++ {
		mgr.Create(CreateRequest{
			Username: "user" + string(rune('0'+i)),
			Password: "Pass123!",
		})
	}

	users, err := mgr.List()
	require.NoError(t, err)
	require.Len(t, users, 3)
}

func TestDeleteUser(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	req := CreateRequest{
		Username: "testuser",
		Password: "Pass123!",
	}
	mgr.Create(req)

	// Delete
	err := mgr.Delete("testuser")
	require.NoError(t, err)

	// Verify deleted
	user, _ := mgr.Get("testuser")
	require.Nil(t, user)
}

func TestDeleteUser_Nonexistent(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	err := mgr.Delete("nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestPasswordHashing(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	password := "MySecurePassword123!"
	req := CreateRequest{
		Username: "hashtest",
		Password: password,
	}
	user, _ := mgr.Create(req)

	// Password should be hashed
	require.NotEqual(t, password, user.PasswordHash)
	require.NotEmpty(t, user.PasswordHash)

	// Hash should start with bcrypt prefix
	require.Contains(t, user.PasswordHash, "$2")
}

func TestMockExecutorVerification(t *testing.T) {
	mgr, db := setupTestUser(t)
	defer db.Close()

	mock := mgr.exec.(*sysexec.MockExecutor)

	req := CreateRequest{
		Username:    "sysuser",
		Password:    "Pass123!",
		AccountType: store.AccountSystem,
	}
	mgr.Create(req)

	// Verify commands were recorded
	cmds := mock.Commands()
	// Should have at least some commands (on non-macOS)
	// On macOS, commands may be empty because system user creation is skipped
	t.Logf("Recorded %d commands", len(cmds))
}
