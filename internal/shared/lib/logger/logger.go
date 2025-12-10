package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lmittmann/tint"
)

// NewLogger creates a new logger with the given log level and optional log file path.
// If logFilePath is empty or invalid, it falls back to stdout.
func NewLogger(logLevel string, logFilePath string) *slog.Logger {
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

	writer, actualLogPath, setupError := getLogWriter(logFilePath)

	var handler slog.Handler
	if writer == os.Stdout {
		handler = tint.NewHandler(writer, &tint.Options{
			Level: level,
		})
	} else {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	switch {
	case actualLogPath != "":
		logger.Info("Logging to file", slog.String("path", actualLogPath))
	case setupError != nil:
		logger.Warn("Failed to setup file logging, using stdout",
			slog.String("requestedPath", logFilePath),
			slog.String("error", setupError.Error()),
		)
	default:
		logger.Info("No log file path configured, using stdout")
	}

	return logger
}

// getLogWriter returns an appropriate io.Writer for logging.
func getLogWriter(logFilePath string) (io.Writer, string, error) {
	if logFilePath == "" {
		return os.Stdout, "", nil
	}

	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return os.Stdout, "", err
	}

	// #nosec G304 -- logFilePath is expected to be user-configurable
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return os.Stdout, "", err
	}

	return file, logFilePath, nil
}
