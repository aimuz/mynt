// Package logger provides structured logging using slog.
package logger

import (
	"log/slog"
	"os"
)

// Level represents log level
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config for logger configuration
type Config struct {
	Level  Level
	Format string // "json" or "text"
}

var defaultLogger *slog.Logger

// Init initializes the global logger with the given configuration
func Init(cfg Config) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns the default logger
func Get() *slog.Logger {
	if defaultLogger == nil {
		// Fallback to default
		defaultLogger = slog.Default()
	}
	return defaultLogger
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message with optional key-value pairs
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}
