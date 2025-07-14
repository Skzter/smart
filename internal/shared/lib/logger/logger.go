package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// NewLogger creates a new logger with the given log level.
func NewLogger(logLevel string) *slog.Logger {
	var level slog.Level

	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level: level,
	}))

	slog.SetDefault(logger)

	return logger
}
