package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user := &User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Email:        "test@example.com",
		AccountType:  AccountVirtual,
		IsAdmin:      false,
		IsActive:     true,
	}

	err := repo.Save(user)
	require.NoError(t, err)
	require.Greater(t, user.ID, int64(0))
	require.NotZero(t, user.CreatedAt)
}

func TestUserRepo_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Save user
	user := &User{
		Username:     "testuser",
		PasswordHash: "hash",
		AccountType:  AccountVirtual,
	}
	repo.Save(user)

	// Get by username
	retrieved, err := repo.GetByUsername("testuser")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, user.Username, retrieved.Username)
	require.Equal(t, user.PasswordHash, retrieved.PasswordHash)
}

func TestUserRepo_GetNonExistent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user, err := repo.GetByUsername("nonexistent")
	require.NoError(t, err)
	require.Nil(t, user)
}

func TestUserRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create multiple users
	users := []*User{
		{Username: "user1", PasswordHash: "hash1", AccountType: AccountVirtual},
		{Username: "user2", PasswordHash: "hash2", AccountType: AccountSystem},
		{Username: "user3", PasswordHash: "hash3", AccountType: AccountVirtual},
	}

	for _, u := range users {
		repo.Save(u)
	}

	// List all users
	list, err := repo.List()
	require.NoError(t, err)
	require.Len(t, list, 3)
}

func TestUserRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create user
	user := &User{
		Username:     "testuser",
		PasswordHash: "hash",
		FullName:     "Original Name",
		IsAdmin:      false,
		AccountType:  AccountVirtual,
	}
	repo.Save(user)

	// Update user
	user.FullName = "Updated Name"
	user.IsAdmin = true

	err := repo.Update(user)
	require.NoError(t, err)

	// Verify update
	retrieved, _ := repo.GetByUsername("testuser")
	require.Equal(t, "Updated Name", retrieved.FullName)
	require.True(t, retrieved.IsAdmin)
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create user
	user := &User{
		Username:     "testuser",
		PasswordHash: "oldhash",
		AccountType:  AccountVirtual,
	}
	repo.Save(user)

	// Update password
	err := repo.UpdatePassword(user.ID, "newhash")
	require.NoError(t, err)

	// Verify
	retrieved, _ := repo.GetByUsername("testuser")
	require.Equal(t, "newhash", retrieved.PasswordHash)
}

func TestUserRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create user
	user := &User{
		Username:     "testuser",
		PasswordHash: "hash",
		AccountType:  AccountVirtual,
	}
	repo.Save(user)

	// Delete user
	err := repo.Delete(user.ID)
	require.NoError(t, err)

	// Verify deleted
	retrieved, _ := repo.GetByUsername("testuser")
	require.Nil(t, retrieved)
}

func TestUserRepo_UniqueUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create first user
	user1 := &User{
		Username:     "duplicate",
		PasswordHash: "hash1",
		AccountType:  AccountVirtual,
	}
	err := repo.Save(user1)
	require.NoError(t, err)

	// Try to create second user with same username
	user2 := &User{
		Username:     "duplicate",
		PasswordHash: "hash2",
		AccountType:  AccountVirtual,
	}
	err = repo.Save(user2)
	require.Error(t, err, "Should not allow duplicate usernames")
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create user
	user := &User{
		Username:     "testuser",
		PasswordHash: "hash",
		AccountType:  AccountVirtual,
	}
	repo.Save(user)

	// Initially no last login
	retrieved, _ := repo.GetByUsername("testuser")
	require.Nil(t, retrieved.LastLogin)

	// Update last login
	err := repo.UpdateLastLogin(user.ID)
	require.NoError(t, err)

	// Verify
	retrieved, _ = repo.GetByUsername("testuser")
	require.NotNil(t, retrieved.LastLogin)
}
