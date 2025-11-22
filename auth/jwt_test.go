package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/store"
)

func TestGenerateToken(t *testing.T) {
	config := DefaultConfig("test-secret-key")
	user := &store.User{
		ID:       1,
		Username: "testuser",
		IsAdmin:  false,
	}

	token, err := GenerateToken(user, config)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	config := DefaultConfig("test-secret-key")
	user := &store.User{
		ID:       1,
		Username: "testuser",
		IsAdmin:  true,
	}

	// Generate token
	token, err := GenerateToken(user, config)
	require.NoError(t, err)

	// Validate token
	claims, err := ValidateToken(token, config)
	require.NoError(t, err)

	require.Equal(t, user.ID, claims.UserID)
	require.Equal(t, user.Username, claims.Username)
	require.Equal(t, user.IsAdmin, claims.IsAdmin)
}

func TestExpiredToken(t *testing.T) {
	config := &Config{
		Secret:        []byte("test-secret"),
		TokenDuration: -1 * time.Hour, // Already expired
	}

	user := &store.User{ID: 1, Username: "test"}

	token, err := GenerateToken(user, config)
	require.NoError(t, err)

	// Should fail validation
	_, err = ValidateToken(token, config)
	require.Error(t, err)
}

func TestInvalidToken(t *testing.T) {
	config := DefaultConfig("test-secret")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not.a.jwt"},
		{"wrong signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, config)
			require.Error(t, err)
		})
	}
}

func TestWrongSecret(t *testing.T) {
	config1 := DefaultConfig("secret1")
	config2 := DefaultConfig("secret2")

	user := &store.User{ID: 1, Username: "test"}

	// Generate with config1
	token, err := GenerateToken(user, config1)
	require.NoError(t, err)

	// Validate with config2 (different secret)
	_, err = ValidateToken(token, config2)
	require.Error(t, err)
}
