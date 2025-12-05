package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/auth"
	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/internal/api"
	"go.aimuz.me/mynt/share"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/task"
	"go.aimuz.me/mynt/user"
	"go.aimuz.me/mynt/zfs"
)

// TestServer creates a test server for integration tests
func setupTestServer(t *testing.T) (*api.Server, *store.DB) {
	// Database
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Components
	pools := zfs.NewManager()
	bus := event.NewBus()
	diskMgr := disk.NewManager()
	tm, _ := task.New(store.NewTaskRepo(db))

	// Share manager
	shareRepo := store.NewShareRepo(db)
	shareMgr := share.NewManager(shareRepo, "")

	// User manager
	userRepo := store.NewUserRepo(db)
	userMgr := user.NewManager(userRepo)

	// Config
	configRepo := store.NewConfigRepo(db)
	jwtSecret, _ := configRepo.GetJWTSecret()
	authConfig := auth.DefaultConfig(jwtSecret)

	// Notification
	notifRepo := store.NewNotificationRepo(db)

	// Snapshot Policy
	snapshotPolicyRepo := store.NewSnapshotPolicyRepo(db)

	// Server (nil for onPolicyChange since we don't have a scheduler in tests)
	srv := api.NewServer(pools, diskMgr, bus, tm, shareMgr, userMgr, configRepo, notifRepo, snapshotPolicyRepo, authConfig, nil)

	return srv, db
}

func TestSetupFlow(t *testing.T) {
	srv, db := setupTestServer(t)

	// 1. Check initial status - should not be initialized
	req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var status map[string]bool
	json.NewDecoder(rr.Body).Decode(&status)
	require.False(t, status["initialized"])

	// 2. Perform setup
	setupData := map[string]string{
		"username":  "admin",
		"password":  "Admin123!",
		"full_name": "Administrator",
		"email":     "admin@example.com",
	}
	body, _ := json.Marshal(setupData)

	req = httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var setupResult map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&setupResult)
	require.NotEmpty(t, setupResult["token"])
	require.NotNil(t, setupResult["user"])

	// 3. Verify system is now initialized
	req = httptest.NewRequest("GET", "/api/v1/setup/status", nil)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	json.NewDecoder(rr.Body).Decode(&status)
	require.True(t, status["initialized"])

	// 4. Try to setup again - should fail
	req = httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)

	// Verify user was created in database
	userRepo := store.NewUserRepo(db)
	admin, err := userRepo.GetByUsername("admin")
	require.NoError(t, err)
	require.NotNil(t, admin)
	require.True(t, admin.IsAdmin)
}

func TestAuthenticationFlow(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Setup first
	setupData := map[string]string{
		"username": "admin",
		"password": "Admin123!",
	}
	body, _ := json.Marshal(setupData)

	req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	// 2. Login with correct credentials
	loginData := map[string]string{
		"username": "admin",
		"password": "Admin123!",
	}
	body, _ = json.Marshal(loginData)

	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var loginResult map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&loginResult)
	token := loginResult["token"].(string)
	require.NotEmpty(t, token)

	// 3. Access protected endpoint with token
	req = httptest.NewRequest("GET", "/api/v1/disks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// 4. Try with wrong password
	wrongLogin := map[string]string{
		"username": "admin",
		"password": "WrongPassword",
	}
	body, _ = json.Marshal(wrongLogin)

	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	// 5. Try protected endpoint without token
	req = httptest.NewRequest("GET", "/api/v1/disks", nil)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAdminEndpoints(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Setup and get admin token
	setupData := map[string]string{
		"username": "admin",
		"password": "Admin123!",
	}
	body, _ := json.Marshal(setupData)
	req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var setupResult map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&setupResult)
	adminToken := setupResult["token"].(string)

	// Create a regular user
	regularUserData := map[string]interface{}{
		"username":     "user",
		"password":     "User123!",
		"account_type": "virtual",
		"is_admin":     false,
	}
	body, _ = json.Marshal(regularUserData)

	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	// Get regular user token
	loginData := map[string]string{
		"username": "user",
		"password": "User123!",
	}
	body, _ = json.Marshal(loginData)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var loginResult map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&loginResult)
	userToken := loginResult["token"].(string)

	// Try to create user as regular user - should fail
	anotherUser := map[string]interface{}{
		"username": "another",
		"password": "Pass123!",
	}
	body, _ = json.Marshal(anotherUser)

	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)

	// Admin can create user
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestCompleteUserJourney(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Setup system
	setupData := map[string]string{
		"username": "admin",
		"password": "Admin123!",
	}
	body, _ := json.Marshal(setupData)
	req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var setupResult map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&setupResult)
	token := setupResult["token"].(string)

	// 2. List disks (may fail without real hardware, that's OK in test)
	req = httptest.NewRequest("GET", "/api/v1/disks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	// Just verify we get a response (may be empty list or error)
	require.NotEqual(t, http.StatusUnauthorized, rr.Code, "Should be authenticated")

	// 3-4. Skip pools and datasets in test environment (no real ZFS)
	// These would require ZFS to be installed and accessible

	// 5. List shares
	req = httptest.NewRequest("GET", "/api/v1/shares", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// 6. List users
	req = httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var users []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&users)
	require.Len(t, users, 1) // Only admin exists

	// 7. List notifications
	req = httptest.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}
