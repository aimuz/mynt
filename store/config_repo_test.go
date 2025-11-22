package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	db, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestConfigRepo_GetSet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	// Test Set
	err := repo.Set("test_key", "test_value")
	require.NoError(t, err)

	// Test Get
	value, err := repo.Get("test_key")
	require.NoError(t, err)
	require.Equal(t, "test_value", value)
}

func TestConfigRepo_GetNonExistent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	_, err := repo.Get("nonexistent")
	require.Error(t, err)
}

func TestConfigRepo_UpdateExisting(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	// Set initial value
	repo.Set("key", "value1")

	// Update value
	err := repo.Set("key", "value2")
	require.NoError(t, err)

	// Verify updated
	value, err := repo.Get("key")
	require.NoError(t, err)
	require.Equal(t, "value2", value)
}

func TestConfigRepo_IsInitialized(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	// Should not be initialized initially
	initialized, err := repo.IsInitialized()
	require.NoError(t, err)
	require.False(t, initialized)

	// Mark as initialized
	err = repo.MarkInitialized()
	require.NoError(t, err)

	// Should now be initialized
	initialized, err = repo.IsInitialized()
	require.NoError(t, err)
	require.True(t, initialized)
}

func TestConfigRepo_GetJWTSecret(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	// First call should generate new secret
	secret1, err := repo.GetJWTSecret()
	require.NoError(t, err)
	require.NotEmpty(t, secret1)

	// Second call should return same secret
	secret2, err := repo.GetJWTSecret()
	require.NoError(t, err)
	require.Equal(t, secret1, secret2)
}

func TestConfigRepo_JWTSecretFormat(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	secret, err := repo.GetJWTSecret()
	require.NoError(t, err)

	// Should be base64 encoded
	require.Greater(t, len(secret), 40, "Secret should be reasonably long")

	// Should not contain spaces
	require.NotContains(t, secret, " ")
}
