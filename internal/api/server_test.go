package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.aimuz.me/mynt/auth"
	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/share"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/sysexec"
	"go.aimuz.me/mynt/task"
	"go.aimuz.me/mynt/user"
	"go.aimuz.me/mynt/zfs"
)

// setupTestServer creates a test server with mocked dependencies.
func setupTestServer(t *testing.T) (*Server, *store.DB) {
	t.Helper()

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	pools := zfs.NewManager()
	bus := event.NewBus()
	diskMgr := disk.NewManager()
	tm, _ := task.New(store.NewTaskRepo(db))

	shareRepo := store.NewShareRepo(db)
	shareMgr := share.NewManager(shareRepo, "")

	userRepo := store.NewUserRepo(db)
	userMgr := user.NewManager(userRepo)
	userMgr.SetExecutor(sysexec.NewMock()) // Mock to avoid real system calls

	configRepo := store.NewConfigRepo(db)
	jwtSecret, _ := configRepo.GetJWTSecret()
	authConfig := auth.DefaultConfig(jwtSecret)

	notifRepo := store.NewNotificationRepo(db)
	snapshotPolicyRepo := store.NewSnapshotPolicyRepo(db)
	diskRepo := store.NewDiskRepo(db)

	srv := NewServer(pools, diskMgr, bus, tm, shareMgr, userMgr, configRepo, notifRepo, snapshotPolicyRepo, diskRepo, authConfig, nil)

	return srv, db
}

// doSetup performs initial system setup and returns the admin token.
func doSetup(t *testing.T, srv *Server) string {
	t.Helper()

	setupData := map[string]string{
		"username": "admin",
		"password": "Admin123!",
	}
	body, _ := json.Marshal(setupData)

	req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("setup failed: status=%d body=%s", rr.Code, rr.Body.String())
	}

	var result map[string]any
	json.NewDecoder(rr.Body).Decode(&result)
	return result["token"].(string)
}

func TestSetupFlow(t *testing.T) {
	srv, db := setupTestServer(t)

	t.Run("initial status not initialized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
		}

		var status map[string]bool
		json.NewDecoder(rr.Body).Decode(&status)
		if status["initialized"] {
			t.Error("expected initialized=false")
		}
	})

	t.Run("perform setup", func(t *testing.T) {
		setupData := map[string]string{
			"username":  "admin",
			"password":  "Admin123!",
			"full_name": "Administrator",
			"email":     "admin@example.com",
		}
		body, _ := json.Marshal(setupData)

		req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status code = %d, want %d, body=%s", rr.Code, http.StatusCreated, rr.Body.String())
		}

		var result map[string]any
		json.NewDecoder(rr.Body).Decode(&result)
		if result["token"] == nil || result["token"] == "" {
			t.Error("expected token in response")
		}
		if result["user"] == nil {
			t.Error("expected user in response")
		}
	})

	t.Run("status after setup", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		var status map[string]bool
		json.NewDecoder(rr.Body).Decode(&status)
		if !status["initialized"] {
			t.Error("expected initialized=true after setup")
		}
	})

	t.Run("setup again fails", func(t *testing.T) {
		setupData := map[string]string{"username": "admin", "password": "Admin123!"}
		body, _ := json.Marshal(setupData)

		req := httptest.NewRequest("POST", "/api/v1/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusForbidden)
		}
	})

	t.Run("user created in db", func(t *testing.T) {
		userRepo := store.NewUserRepo(db)
		admin, err := userRepo.GetByUsername("admin")
		if err != nil {
			t.Fatalf("GetByUsername: %v", err)
		}
		if admin == nil {
			t.Fatal("admin user not found")
		}
		if !admin.IsAdmin {
			t.Error("expected IsAdmin=true")
		}
	})
}

func TestAuthenticationFlow(t *testing.T) {
	srv, _ := setupTestServer(t)
	doSetup(t, srv)

	t.Run("login success", func(t *testing.T) {
		loginData := map[string]string{"username": "admin", "password": "Admin123!"}
		body, _ := json.Marshal(loginData)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
		}

		var result map[string]any
		json.NewDecoder(rr.Body).Decode(&result)
		if result["token"] == nil || result["token"] == "" {
			t.Error("expected token in response")
		}
	})

	t.Run("login wrong password", func(t *testing.T) {
		loginData := map[string]string{"username": "admin", "password": "WrongPassword"}
		body, _ := json.Marshal(loginData)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("protected endpoint without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/disks", nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("protected endpoint with token", func(t *testing.T) {
		// Login first
		loginData := map[string]string{"username": "admin", "password": "Admin123!"}
		body, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		var result map[string]any
		json.NewDecoder(rr.Body).Decode(&result)
		token := result["token"].(string)

		// Access protected endpoint
		req = httptest.NewRequest("GET", "/api/v1/disks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr = httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

func TestAdminEndpoints(t *testing.T) {
	srv, _ := setupTestServer(t)
	adminToken := doSetup(t, srv)

	t.Run("admin can create user", func(t *testing.T) {
		userData := map[string]any{
			"username":     "testuser",
			"password":     "User123!",
			"account_type": "virtual",
			"is_admin":     false,
		}
		body, _ := json.Marshal(userData)

		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status code = %d, want %d, body=%s", rr.Code, http.StatusCreated, rr.Body.String())
		}
	})

	t.Run("regular user cannot create user", func(t *testing.T) {
		// Login as regular user
		loginData := map[string]string{"username": "testuser", "password": "User123!"}
		body, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		var result map[string]any
		json.NewDecoder(rr.Body).Decode(&result)
		userToken := result["token"].(string)

		// Try to create another user
		newUser := map[string]any{"username": "another", "password": "Pass123!"}
		body, _ = json.Marshal(newUser)
		req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		rr = httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusForbidden)
		}
	})
}

func TestAPIEndpoints(t *testing.T) {
	srv, _ := setupTestServer(t)
	token := doSetup(t, srv)

	endpoints := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/api/v1/shares", http.StatusOK},
		{"GET", "/api/v1/users", http.StatusOK},
		{"GET", "/api/v1/notifications", http.StatusOK},
	}

	for _, e := range endpoints {
		t.Run(e.method+" "+e.path, func(t *testing.T) {
			req := httptest.NewRequest(e.method, e.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != e.want {
				t.Errorf("status code = %d, want %d", rr.Code, e.want)
			}
		})
	}
}
