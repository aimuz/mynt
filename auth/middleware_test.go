package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/store"
)

func TestRequireAuth(t *testing.T) {
	config := DefaultConfig("test-secret")
	middleware := NewMiddleware(config)

	user := &store.User{
		ID:       1,
		Username: "testuser",
		IsAdmin:  false,
	}
	token, _ := GenerateToken(user, config)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectClaims   bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectClaims:   true,
		},
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectClaims:   false,
		},
		{
			name:           "invalid format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectClaims:   false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectClaims:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := GetUserClaims(r.Context())
				if tt.expectClaims {
					require.NotNil(t, claims)
					require.Equal(t, user.Username, claims.Username)
				}
				w.WriteHeader(http.StatusOK)
			}))

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	config := DefaultConfig("test-secret")
	middleware := NewMiddleware(config)

	adminUser := &store.User{ID: 1, Username: "admin", IsAdmin: true}
	regularUser := &store.User{ID: 2, Username: "user", IsAdmin: false}

	adminToken, _ := GenerateToken(adminUser, config)
	userToken, _ := GenerateToken(regularUser, config)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "admin user",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "regular user",
			authHeader:     "Bearer " + userToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no auth",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireAuth(
				middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})),
			)

			req := httptest.NewRequest("GET", "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestOptionalAuth(t *testing.T) {
	config := DefaultConfig("test-secret")
	middleware := NewMiddleware(config)

	user := &store.User{ID: 1, Username: "test"}
	token, _ := GenerateToken(user, config)

	tests := []struct {
		name         string
		authHeader   string
		expectClaims bool
	}{
		{
			name:         "with valid token",
			authHeader:   "Bearer " + token,
			expectClaims: true,
		},
		{
			name:         "without token",
			authHeader:   "",
			expectClaims: false,
		},
		{
			name:         "with invalid token",
			authHeader:   "Bearer invalid",
			expectClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := GetUserClaims(r.Context())
				if tt.expectClaims {
					require.NotNil(t, claims)
				} else {
					require.Nil(t, claims)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestGetUserClaims(t *testing.T) {
	config := DefaultConfig("test-secret")
	user := &store.User{ID: 1, Username: "test", IsAdmin: true}
	token, _ := GenerateToken(user, config)

	claims, err := ValidateToken(token, config)
	require.NoError(t, err)

	// Create context with claims
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	// Get claims from context
	retrieved := GetUserClaims(ctx)
	require.NotNil(t, retrieved)
	require.Equal(t, user.ID, retrieved.UserID)
	require.Equal(t, user.Username, retrieved.Username)
	require.Equal(t, user.IsAdmin, retrieved.IsAdmin)

	// Test empty context
	emptyCtx := context.Background()
	nilClaims := GetUserClaims(emptyCtx)
	require.Nil(t, nilClaims)
}
