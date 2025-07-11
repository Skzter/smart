package lib

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// NewLogger creates a new logger with the given log level.
func NewLogger(logLevel string) *slog.Logger {
	level := slog.LevelInfo

	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level: level,
	}))

	slog.SetDefault(logger)

	return logger
}
