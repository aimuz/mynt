package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.aimuz.me/mynt/auth"
	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/share"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/task"
	"go.aimuz.me/mynt/user"
	webui "go.aimuz.me/mynt/web-ui"
	"go.aimuz.me/mynt/zfs"
)

// Server represents the HTTP API server.
type Server struct {
	zfs          *zfs.Manager
	disk         *disk.Manager
	bus          *event.Bus
	tm           *task.Manager
	share        *share.Manager
	user         *user.Manager
	config       *store.ConfigRepo
	notification *store.NotificationRepo
	authConfig   *auth.Config
	authMw       *auth.Middleware
	mux          *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(zfs *zfs.Manager, diskMgr *disk.Manager, bus *event.Bus, tm *task.Manager, sm *share.Manager, um *user.Manager, cfg *store.ConfigRepo, notif *store.NotificationRepo, authCfg *auth.Config) *Server {
	s := &Server{
		zfs:          zfs,
		disk:         diskMgr,
		bus:          bus,
		tm:           tm,
		share:        sm,
		user:         um,
		config:       cfg,
		notification: notif,
		authConfig:   authCfg,
		authMw:       auth.NewMiddleware(authCfg),
		mux:          http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Static Files (public)
	s.mux.Handle("/", spaHandler(webui.FS, "index.html"))

	// Setup route (only available if not initialized)
	s.mux.HandleFunc("POST /api/v1/setup", s.handleSetup)
	s.mux.HandleFunc("GET /api/v1/setup/status", s.handleSetupStatus)

	// Public routes (no auth required)
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)

	// Protected API routes - all require authentication
	// Apply auth middleware to all /api/v1/ routes except auth
	s.mux.HandleFunc("GET /api/v1/disks", s.protected(s.handleListDisks))
	s.mux.HandleFunc("GET /api/v1/disks/smart", s.protected(s.handleDiskSmart))
	s.mux.HandleFunc("GET /api/v1/pools", s.protected(s.handleListPools))
	s.mux.HandleFunc("POST /api/v1/pools", s.protected(s.handleCreatePool))

	// Enhanced pool operations
	s.mux.HandleFunc("POST /api/v1/pools/{name}/scrub", s.protected(s.handlePoolScrub))
	s.mux.HandleFunc("GET /api/v1/pools/{name}/scrub/status", s.protected(s.handleScrubStatus))

	s.mux.HandleFunc("GET /api/v1/datasets", s.protected(s.handleListDatasets))
	s.mux.HandleFunc("POST /api/v1/datasets", s.protected(s.handleCreateDataset))
	s.mux.HandleFunc("GET /api/v1/datasets/{name...}", s.protected(s.handleGetDataset))
	s.mux.HandleFunc("DELETE /api/v1/datasets/{name...}", s.protected(s.handleDestroyDataset))
	s.mux.HandleFunc("PUT /api/v1/datasets/{name...}/quota", s.protected(s.handleSetDatasetQuota))

	// Snapshot endpoints
	s.mux.HandleFunc("GET /api/v1/snapshots", s.protected(s.handleListSnapshots))
	s.mux.HandleFunc("POST /api/v1/snapshots", s.protected(s.handleCreateSnapshot))
	s.mux.HandleFunc("DELETE /api/v1/snapshots/{name...}", s.protected(s.handleDestroySnapshot))
	s.mux.HandleFunc("POST /api/v1/snapshots/{name...}/rollback", s.protected(s.handleRollbackSnapshot))

	// Shares
	s.mux.HandleFunc("GET /api/v1/shares", s.protected(s.handleListShares))
	s.mux.HandleFunc("POST /api/v1/shares", s.protected(s.handleCreateShare))
	s.mux.HandleFunc("DELETE /api/v1/shares/{id}", s.protected(s.handleDeleteShare))

	// Users (admin only for create/delete)
	s.mux.HandleFunc("GET /api/v1/users", s.protected(s.handleListUsers))
	s.mux.HandleFunc("POST /api/v1/users", s.adminOnly(s.handleCreateUser))
	s.mux.HandleFunc("DELETE /api/v1/users/{username}", s.adminOnly(s.handleDeleteUser))

	// Notifications
	s.mux.HandleFunc("GET /api/v1/notifications", s.protected(s.handleListNotifications))
	s.mux.HandleFunc("POST /api/v1/notifications/{id}/read", s.protected(s.handleMarkRead))
	s.mux.HandleFunc("POST /api/v1/notifications/{id}/ack", s.protected(s.handleMarkAcknowledged))
	s.mux.HandleFunc("DELETE /api/v1/notifications/{id}", s.protected(s.handleDeleteNotification))
	s.mux.HandleFunc("GET /api/v1/notifications/count", s.protected(s.handleCountNotifications))

	// Real-time events - SSE
	s.mux.HandleFunc("GET /api/v1/events", s.protected(s.handleEvents))
}

// protected wraps a handler with authentication requirement.
func (s *Server) protected(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.authMw.RequireAuth(handler).ServeHTTP(w, r)
	}
}

// adminOnly wraps a handler with admin authentication requirement.
func (s *Server) adminOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.authMw.RequireAuth(s.authMw.RequireAdmin(handler)).ServeHTTP(w, r)
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Setup handlers

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	initialized, err := s.config.IsInitialized()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{
		"initialized": initialized,
	})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	// Check if already initialized
	initialized, err := s.config.IsInitialized()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if initialized {
		http.Error(w, "system already initialized", http.StatusForbidden)
		return
	}

	// Parse request
	var req user.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Force admin and system account
	req.IsAdmin = true
	req.AccountType = "system"

	// Create admin user
	admin, err := s.user.Create(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mark system as initialized
	if err := s.config.MarkInitialized(); err != nil {
		http.Error(w, "failed to mark initialized", http.StatusInternalServerError)
		return
	}

	// Generate token for immediate login
	token, err := auth.GenerateToken(admin, s.authConfig)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token and user
	// Return token and user
	respondJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  admin,
	})
}

// Authentication handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Verify credentials
	user, err := s.user.VerifyPassword(req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user, s.authConfig)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token and user info
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// Resource handlers

func (s *Server) handleListDisks(w http.ResponseWriter, r *http.Request) {
	disks, err := s.disk.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, disks)
}

func (s *Server) handleDiskSmart(w http.ResponseWriter, r *http.Request) {
	diskName := r.URL.Query().Get("name")
	if diskName == "" {
		http.Error(w, "disk name required", http.StatusBadRequest)
		return
	}

	report, err := disk.Smart(r.Context(), diskName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleListPools(w http.ResponseWriter, r *http.Request) {
	pools, err := s.zfs.ListPools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, pools)
}

func (s *Server) handleCreatePool(w http.ResponseWriter, r *http.Request) {
	var req zfs.CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || len(req.Devices) == 0 {
		http.Error(w, "name and devices are required", http.StatusBadRequest)
		return
	}

	if err := s.zfs.CreatePool(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := s.zfs.ListDatasets(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, datasets)
}

func (s *Server) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	var req zfs.CreateDatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.zfs.CreateDataset(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleGetDataset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "dataset name required", http.StatusBadRequest)
		return
	}

	dataset, err := s.zfs.GetDataset(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, dataset)
}

func (s *Server) handleDestroyDataset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "dataset name required", http.StatusBadRequest)
		return
	}

	if err := s.zfs.DestroyDataset(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Share handlers

func (s *Server) handleListShares(w http.ResponseWriter, r *http.Request) {
	protocol := r.URL.Query().Get("protocol")

	shares, err := s.share.ListShares(protocol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, shares)
}

func (s *Server) handleCreateShare(w http.ResponseWriter, r *http.Request) {
	var share store.Share
	if err := json.NewDecoder(r.Body).Decode(&share); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Default to SMB if not specified
	if share.Protocol == "" {
		share.Protocol = "smb"
	}

	if err := s.share.CreateShare(&share); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, share)
}

func (s *Server) handleDeleteShare(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.share.DeleteShare(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// User handlers

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.user.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req user.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.user.Create(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	if err := s.user.Delete(username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListNotifications returns notification history with filtering.
func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := store.NotificationStatus(r.URL.Query().Get("status"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	notifications, err := s.notification.List(status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, notifications)
}

// handleMarkRead marks a notification as read.
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.notification.MarkRead(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleMarkAcknowledged marks a notification as acknowledged (processed).
func (s *Server) handleMarkAcknowledged(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.notification.MarkAcknowledged(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteNotification deletes a notification.
func (s *Server) handleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.notification.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Pool scrub handlers

func (s *Server) handlePoolScrub(w http.ResponseWriter, r *http.Request) {
	poolName := r.PathValue("name")
	if poolName == "" {
		http.Error(w, "pool name required", http.StatusBadRequest)
		return
	}

	var req struct {
		Action string `json:"action"` // start, stop, pause
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Only "start" is supported for now via the Scrub method
	if req.Action != "start" {
		http.Error(w, "only 'start' action is supported", http.StatusBadRequest)
		return
	}

	if err := s.zfs.Scrub(r.Context(), poolName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleScrubStatus(w http.ResponseWriter, r *http.Request) {
	poolName := r.PathValue("name")
	if poolName == "" {
		http.Error(w, "pool name required", http.StatusBadRequest)
		return
	}

	status, err := s.zfs.ScrubStatus(r.Context(), poolName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": status,
	})
}

// Dataset quota handler

func (s *Server) handleSetDatasetQuota(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "dataset name required", http.StatusBadRequest)
		return
	}

	var req struct {
		Quota uint64 `json:"quota"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.zfs.SetQuota(r.Context(), name, req.Quota); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Snapshot handlers

func (s *Server) handleListSnapshots(w http.ResponseWriter, r *http.Request) {
	datasetName := r.URL.Query().Get("dataset")
	if datasetName == "" {
		http.Error(w, "dataset parameter required", http.StatusBadRequest)
		return
	}

	snapshots, err := s.zfs.ListSnapshots(r.Context(), datasetName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, snapshots)
}

func (s *Server) handleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var req zfs.CreateSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	snapshot, err := s.zfs.CreateSnapshot(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, snapshot)
}

func (s *Server) handleDestroySnapshot(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "snapshot name required", http.StatusBadRequest)
		return
	}

	if err := s.zfs.DestroySnapshot(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRollbackSnapshot(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "snapshot name required", http.StatusBadRequest)
		return
	}

	if err := s.zfs.RollbackSnapshot(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCountNotifications returns notification counts by status.
func (s *Server) handleCountNotifications(w http.ResponseWriter, r *http.Request) {
	unread, _ := s.notification.Count(store.NotificationUnread)
	read, _ := s.notification.Count(store.NotificationRead)
	acked, _ := s.notification.Count(store.NotificationAcked)

	respondJSON(w, http.StatusOK, map[string]int{
		"unread":       unread,
		"read":         read,
		"acknowledged": acked,
		"total":        unread + read + acked,
	})
}

// handleEvents provides Server-Sent Events for real-time notifications.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Subscribe to all events
	ch := s.bus.Subscribe("*")
	defer func() {
		s.bus.Unsubscribe("*", ch)
	}()

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().Unix())
	w.(http.Flusher).Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-ch:
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
	}
}
