// Package logger provides logging functionality for the URL shortener service.
package logger

import (
	"log/slog"
	"os"
)

// Log is a global variable that holds the logger instance.
var Log *slog.Logger

// NewLogger is a function that creates a new logger instance.
func NewLogger(level string) error {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	Log = slog.New(handler)

	return nil
}
