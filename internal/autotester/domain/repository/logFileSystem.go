package repository

import (
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// LogFileSystem is seperate Filesystem for logs
type LogFileSystem FileSystem

// NewLogFileSystem creates a filesystem for log files
func NewLogFileSystem(logDir string) (LogFileSystem, error) {
	if err := assert.StringNotEmpty(logDir); err != nil {
		return nil, fmt.Errorf("root must not be empty")
	}
	fs, err := NewOSFileSystem(logDir)
	return LogFileSystem(fs), err
}
