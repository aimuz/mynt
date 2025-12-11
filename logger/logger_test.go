package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInit_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	
	// Save original default
	orig := slog.Default()
	defer slog.SetDefault(orig)

	cfg := Config{
		Level:  LevelInfo,
		Format: "text",
	}

	// Redirect output for testing
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: cfg.Level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message, got: %s", output)
	}
	if !strings.Contains(output, "key") {
		t.Errorf("output should contain key, got: %s", output)
	}
}

func TestInit_JSONFormat(t *testing.T) {
	var buf bytes.Buffer

	// Save original default
	orig := slog.Default()
	defer slog.SetDefault(orig)

	cfg := Config{
		Level:  LevelDebug,
		Format: "json",
	}

	// Redirect output for testing
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: cfg.Level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Debug("test debug", "num", 42)

	output := buf.String()
	if !strings.Contains(output, "test debug") {
		t.Errorf("output should contain message, got: %s", output)
	}
	if !strings.Contains(output, "num") {
		t.Errorf("output should contain key, got: %s", output)
	}
	// JSON format should have structure
	if !strings.Contains(output, "{") {
		t.Errorf("JSON output should contain braces, got: %s", output)
	}
}

func TestLevels(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  slog.Level
	}{
		{
			name:  "debug",
			level: LevelDebug,
			want:  slog.LevelDebug,
		},
		{
			name:  "info",
			level: LevelInfo,
			want:  slog.LevelInfo,
		},
		{
			name:  "warn",
			level: LevelWarn,
			want:  slog.LevelWarn,
		},
		{
			name:  "error",
			level: LevelError,
			want:  slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level != tt.want {
				t.Errorf("level = %v, want %v", tt.level, tt.want)
			}
		})
	}
}

func TestGet_WithoutInit(t *testing.T) {
	// Reset defaultLogger
	defaultLogger = nil

	// Should return a default logger without panicking
	logger := Get()
	if logger == nil {
		t.Error("Get() should return a logger, got nil")
	}
}

func TestGet_AfterInit(t *testing.T) {
	// Save original
	orig := defaultLogger
	defer func() { defaultLogger = orig }()

	cfg := Config{
		Level:  LevelInfo,
		Format: "text",
	}
	Init(cfg)

	logger := Get()
	if logger == nil {
		t.Error("Get() should return initialized logger, got nil")
	}
	if logger != defaultLogger {
		t.Error("Get() should return the same logger as defaultLogger")
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer

	// Save original
	orig := slog.Default()
	defer slog.SetDefault(orig)

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Debug("debug message", "id", 123)

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("output should contain debug message, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer

	// Save original
	orig := slog.Default()
	defer slog.SetDefault(orig)

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: LevelInfo})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Info("info message", "status", "ok")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("output should contain info message, got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer

	// Save original
	orig := slog.Default()
	defer slog.SetDefault(orig)

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: LevelWarn})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Warn("warning message", "code", 404)

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("output should contain warning message, got: %s", output)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer

	// Save original
	orig := slog.Default()
	defer slog.SetDefault(orig)

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: LevelError})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultLogger = logger

	Error("error message", "code", 500)

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("output should contain error message, got: %s", output)
	}
}

func TestLevelFiltering(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		logFunc    func()
		shouldLog  bool
	}{
		{
			name:  "debug_level_logs_debug",
			level: LevelDebug,
			logFunc: func() {
				Debug("debug")
			},
			shouldLog: true,
		},
		{
			name:  "info_level_filters_debug",
			level: LevelInfo,
			logFunc: func() {
				Debug("debug")
			},
			shouldLog: false,
		},
		{
			name:  "info_level_logs_info",
			level: LevelInfo,
			logFunc: func() {
				Info("info")
			},
			shouldLog: true,
		},
		{
			name:  "error_level_filters_info",
			level: LevelError,
			logFunc: func() {
				Info("info")
			},
			shouldLog: false,
		},
		{
			name:  "error_level_logs_error",
			level: LevelError,
			logFunc: func() {
				Error("error")
			},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Save original
			orig := slog.Default()
			defer slog.SetDefault(orig)

			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: tt.level})
			logger := slog.New(handler)
			slog.SetDefault(logger)
			defaultLogger = logger

			tt.logFunc()

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldLog {
				t.Errorf("shouldLog = %v, but hasOutput = %v, output: %s", tt.shouldLog, hasOutput, output)
			}
		})
	}
}
