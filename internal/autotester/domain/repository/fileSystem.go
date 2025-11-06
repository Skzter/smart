package repository

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Default permissions used by the filesystem implementation.
const (
	DefaultDirPerm  os.FileMode = 0o755
	DefaultFilePerm os.FileMode = 0o644
)

// FileSystem is a minimal filesystem abstraction for repository I/O.
type FileSystem interface {
	// MkdirAll creates a directory path (including parents) inside the
	// configured root using the package default directory permissions.
	MkdirAll(path string) error

	// WriteFile writes data to relativeFilePath inside the configured root. The
	// operation is performed atomically: data is written to a temporary
	// file and then renamed to the final destination. The package default
	// file permission is used. The caller MUST ensure that the parent
	// directories exist (for example by calling MkdirAll) — this method
	// will NOT create parent directories.
	WriteFile(relativeFilePath string, data []byte) error

	// ReadFile reads and returns the content of the named file inside the
	// configured root.
	ReadFile(relativeFilePath string) ([]byte, error)

	// ReadDir lists the non-directory entries directly under the provided
	// path inside the configured root and returns their base names. The
	// returned names do not include the directory path.
	ReadDir(path string) ([]string, error)

	// Remove removes the named path from the configured root.
	// If recursive is true the implementation MUST remove the path and all
	// children (recursive delete). If recursive is false the implementation
	// should remove only the named path and fail for non-empty directories.
	//
	// If the path does not exist the method
	// returns nil. Implementations must reject operations that would escape
	// the configured root.
	Remove(path string, recursive bool) error

	// Stat returns file information for the given path relative to the filesystem root.
	// Returns an error if the path would escape the configured Root or if the file does not exist.
	GetFileStats(path string) (os.FileInfo, error)

	// GetValidatedPath validates a relative path and returns it if it's within the filesystem root.
	// Returns an error if the path would escape the configured Root or if validation fails.
	GetValidatedPath(relativePath string) (string, error)
}

// osFileSystem is a implementation of the FileSystem interface that confines all operations
// to a configurable Root directory. Use NewOSFileSystem(root) to obtain an
// instance. Operations that would escape the Root are rejected with an error.
type osFileSystem struct {
	Root string
}

// NewOSFileSystem returns an osFileSystem rooted at the provided directory.
// The root is created if it does not exist. Root must not be empty or absolute.
func NewOSFileSystem(root string) (FileSystem, error) {
	if err := assert.StringNotEmpty(root); err != nil {
		return nil, fmt.Errorf("root must not be empty")
	}
	if path.IsAbs(root) {
		return nil, fmt.Errorf("absolute paths are not allowed: %s", root)
	}
	if err := os.MkdirAll(root, DefaultDirPerm); err != nil {
		return nil, fmt.Errorf("create root: %w", err)
	}
	return &osFileSystem{
		Root: filepath.Clean(root),
	}, nil
}

func (fs *osFileSystem) MkdirAll(path string) error {
	fullPath, err := fs.validateAndGetFullPath(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, DefaultDirPerm)
}

func (fs *osFileSystem) WriteFile(relativeFilePath string, data []byte) error {
	if err := assert.StringNotEmpty(relativeFilePath); err != nil {
		return fmt.Errorf("relativeFilePath must be not empty")
	}
	fullPath, err := fs.validateAndGetFullPath(relativeFilePath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpName, DefaultFilePerm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpName, fullPath); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

func (fs *osFileSystem) ReadFile(relativeFilePath string) ([]byte, error) {
	if err := assert.StringNotEmpty(relativeFilePath); err != nil {
		return nil, fmt.Errorf("relativeFilePath must be not empty")
	}
	fullPath, err := fs.validateAndGetFullPath(relativeFilePath)
	if err != nil {
		return nil, err
	}

	resolvedPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	if rp, err := filepath.EvalSymlinks(resolvedPath); err == nil {
		resolvedPath = rp
	}

	if !fs.isUnderRoot(resolvedPath) {
		return nil, fmt.Errorf("path escapes root: %s", relativeFilePath)
	}
	return os.ReadFile(resolvedPath)
}

func (fs *osFileSystem) ReadDir(path string) ([]string, error) {
	fullPath, err := fs.validateAndGetFullPath(path)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (fs *osFileSystem) Remove(path string, recursive bool) error {
	full, err := fs.validateAndGetFullPath(path)
	if err != nil {
		return err
	}

	if recursive {
		err = os.RemoveAll(full)
	} else {
		err = os.Remove(full)
	}

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (fs *osFileSystem) GetFileStats(path string) (os.FileInfo, error) {
	full, err := fs.validateAndGetFullPath(path)
	if err != nil {
		return nil, err
	}

	return os.Lstat(full)
}

func (fs *osFileSystem) GetValidatedPath(relativePath string) (string, error) {
	return fs.validateAndGetFullPath(relativePath)
}

func (fs *osFileSystem) fullPath(p string) string {
	return filepath.Clean(filepath.Join(fs.Root, p))
}

// validateAndGetFullPath validates that the given path is relative and does not escape
// the filesystem root. Returns the full path if valid, otherwise returns an error.
func (fs *osFileSystem) validateAndGetFullPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be relative to root: %s", path)
	}
	fullPath := fs.fullPath(path)
	if !fs.isUnderRoot(fullPath) {
		return "", fmt.Errorf("path escapes root: %s", path)
	}
	return fullPath, nil
}

// isUnderRoot reports whether the given path is located inside the
// filesystem root. It resolves both the root and the path to absolute paths,
// follows symlinks using filepath.EvalSymlinks, and then checks if the resolved
// path has the resolved root as a prefix. Returns true if the path is under
// the root or equal to the root, false otherwise.
func (fs *osFileSystem) isUnderRoot(path string) bool {
	rootAbs, err := filepath.Abs(fs.Root)
	if err != nil {
		return false
	}
	if r, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootAbs = r
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	if p, err := filepath.EvalSymlinks(pathAbs); err == nil {
		pathAbs = p
	}

	return strings.HasPrefix(pathAbs, rootAbs+string(os.PathSeparator)) || pathAbs == rootAbs
}
