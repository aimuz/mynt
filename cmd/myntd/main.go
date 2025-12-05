package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.aimuz.me/mynt/auth"
	"go.aimuz.me/mynt/disk"
	"go.aimuz.me/mynt/event"
	"go.aimuz.me/mynt/internal/api"
	"go.aimuz.me/mynt/logger"
	"go.aimuz.me/mynt/monitor"
	"go.aimuz.me/mynt/scheduler"
	"go.aimuz.me/mynt/share"
	"go.aimuz.me/mynt/store"
	"go.aimuz.me/mynt/task"
	"go.aimuz.me/mynt/user"
	"go.aimuz.me/mynt/zfs"
)

func main() {
	// Flags
	dbPath := flag.String("db", "mynt.db", "Path to SQLite database")
	addr := flag.String("addr", ":8080", "HTTP API address")
	smbConfig := flag.String("smb-config", "", "Path to smb.conf (empty for auto-detect)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", "text", "Log format (text, json)")
	enableLoopDevices := flag.Bool("enable-loop-devices", false, "Enable detection of loop devices (for testing)")
	flag.Parse()

	// Initialize logger
	level := logger.LevelInfo
	switch *logLevel {
	case "debug":
		level = logger.LevelDebug
	case "warn":
		level = logger.LevelWarn
	case "error":
		level = logger.LevelError
	}

	logger.Init(logger.Config{
		Level:  level,
		Format: *logFormat,
	})

	// Database
	db, err := store.Open(*dbPath)
	if err != nil {
		logger.Error("failed to open database", "path", *dbPath, "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Config repository
	configRepo := store.NewConfigRepo(db)

	// Get or generate JWT secret
	jwtSecret, err := configRepo.GetJWTSecret()
	if err != nil {
		logger.Error("failed to get JWT secret", "error", err)
		os.Exit(1)
	}

	// Task manager
	mgr, err := task.New(store.NewTaskRepo(db))
	if err != nil {
		logger.Error("failed to initialize task manager", "error", err)
		os.Exit(1)
	}

	// ZFS
	pools := zfs.NewManager()

	// Event bus with persistence
	bus := event.NewBus()
	notificationRepo := store.NewNotificationRepo(db)
	snapshotPolicyRepo := store.NewSnapshotPolicyRepo(db)
	bus.SetPersister(notificationRepo)

	// Share manager
	shareRepo := store.NewShareRepo(db)
	shareMgr := share.NewManager(shareRepo, *smbConfig)

	// User manager
	userRepo := store.NewUserRepo(db)
	userMgr := user.NewManager(userRepo)

	// Auth config
	authConfig := auth.DefaultConfig(jwtSecret)

	// Monitoring with disk repository
	diskRepo := store.NewDiskRepo(db)

	// Disk Manager
	var diskOpts []disk.ManagerOption
	if *enableLoopDevices {
		diskOpts = append(diskOpts, disk.WithLoopDevices())
	}
	diskMgr := disk.NewManager(diskOpts...)

	diskScanner := monitor.NewDiskScanner(bus, diskRepo, diskMgr)
	zfsScanner := monitor.NewZFSScanner(bus, pools)
	mon := monitor.New(
		[]monitor.Scanner{diskScanner, zfsScanner},
		30*time.Second,
	)

	ctx := context.Background()
	mon.Start(ctx)
	defer mon.Stop()

	logger.Info("monitoring started", "scanners", 2, "interval_sec", 30)

	// Snapshot Policy Scheduler
	snapshotScheduler := scheduler.New(snapshotPolicyRepo, pools)
	if err := snapshotScheduler.Start(ctx); err != nil {
		logger.Error("failed to start snapshot scheduler", "error", err)
		os.Exit(1)
	}
	defer snapshotScheduler.Stop()

	// Check initialization status
	initialized, _ := configRepo.IsInitialized()
	if !initialized {
		logger.Warn("system not initialized",
			"setup_url", "http://localhost:8080/setup")
	}

	// API Server with authentication
	srv := api.NewServer(pools, diskMgr, bus, mgr, shareMgr, userMgr, configRepo, notificationRepo, snapshotPolicyRepo, authConfig, func() { _ = snapshotScheduler.Reload() })
	httpSrv := &http.Server{
		Addr:    *addr,
		Handler: srv,
	}

	// Start server
	go func() {
		logger.Info("starting http server", "address", *addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
