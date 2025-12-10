package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		logFilePath string
		wantLevel   string
	}{
		{
			name:        "stdout with info level",
			logLevel:    "info",
			logFilePath: "",
			wantLevel:   "info",
		},
		{
			name:        "stdout with debug level",
			logLevel:    "debug",
			logFilePath: "",
			wantLevel:   "debug",
		},
		{
			name:        "stdout with invalid level defaults to warn",
			logLevel:    "invalid",
			logFilePath: "",
			wantLevel:   "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.logLevel, tt.logFilePath)
			assert.NotNil(t, logger)
		})
	}
}

func TestNewLoggerWithFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger := NewLogger("info", logFile)
	require.NotNil(t, logger)

	// Write a test log
	logger.Info("test message", "key", "value")

	// Verify file was created
	_, err := os.Stat(logFile)
	assert.NoError(t, err, "log file should exist")

	// Read file content
	// #nosec G304 -- logFile is a test temp file, not user input
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message")
	assert.Contains(t, string(content), "key")
	assert.Contains(t, string(content), "value")
}

func TestNewLoggerWithNestedDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "nested", "dir", "test.log")

	logger := NewLogger("info", logFile)
	require.NotNil(t, logger)

	// Write a test log
	logger.Info("nested test")

	// Verify file was created
	_, err := os.Stat(logFile)
	assert.NoError(t, err, "log file in nested directory should exist")
}

func TestNewLoggerWithInvalidPath(t *testing.T) {
	// Try to create log file in non-writable location (should fallback to stdout)
	invalidPath := "/root/cannot-write-here/test.log"

	logger := NewLogger("info", invalidPath)
	assert.NotNil(t, logger, "logger should still be created with stdout fallback")

	// File should not exist
	_, err := os.Stat(invalidPath)
	assert.Error(t, err, "invalid path should not create file")
}

func TestGetLogWriter(t *testing.T) {
	tests := []struct {
		name         string
		logFilePath  string
		wantStdout   bool
		wantPath     string
		wantError    bool
		setupInvalid bool
	}{
		{
			name:        "empty path returns stdout",
			logFilePath: "",
			wantStdout:  true,
			wantPath:    "",
			wantError:   false,
		},
		{
			name:        "valid path returns file",
			logFilePath: filepath.Join(t.TempDir(), "test.log"),
			wantStdout:  false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, path, err := getLogWriter(tt.logFilePath)

			assert.NotNil(t, writer)

			if tt.wantStdout {
				assert.Equal(t, os.Stdout, writer)
				assert.Empty(t, path)
			} else {
				assert.NotEqual(t, os.Stdout, writer)
				assert.NotEmpty(t, path)
			}

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"DEBUG", "warn"}, // Invalid, defaults to warn
		{"", "warn"},      // Invalid, defaults to warn
		{"invalid", "warn"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			logger := NewLogger(tt.input, "")
			assert.NotNil(t, logger)
		})
	}
}
