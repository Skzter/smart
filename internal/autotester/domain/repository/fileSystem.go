package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// WriteFile writes data to filename inside the configured root. The
	// operation is performed atomically: data is written to a temporary
	// file and then renamed to the final destination. The package default
	// file permission is used. The caller MUST ensure that the parent
	// directories exist (for example by calling MkdirAll) — this method
	// will NOT create parent directories.
	WriteFile(filename string, data []byte) error

	// ReadFile reads and returns the content of the named file inside the
	// configured root.
	ReadFile(filename string) ([]byte, error)

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
}

// osFileSystem is a implementation of the FileSystem interface that confines all operations
// to a configurable Root directory. Use NewOSFileSystem(root) to obtain an
// instance. Operations that would escape the Root are rejected with an error.
type osFileSystem struct {
	Root string
}

// NewOSFileSystem returns an osFileSystem rooted at the provided directory.
// The root is created if it does not exist. Root must not be empty.
func NewOSFileSystem(root string) (FileSystem, error) {
	if root == "" {
		return nil, fmt.Errorf("root must not be empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	if err := os.MkdirAll(abs, DefaultDirPerm); err != nil {
		return nil, fmt.Errorf("create root: %w", err)
	}
	return &osFileSystem{
		Root: filepath.Clean(abs),
	}, nil
}

func (fs osFileSystem) MkdirAll(path string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative to root: %s", path)
	}

	fullPath := fs.fullPath(path)
	if !fs.isUnderRoot(fullPath) {
		return fmt.Errorf("path escapes root: %s", path)
	}
	return os.MkdirAll(fullPath, DefaultDirPerm)
}

// WriteFile writes data to filename inside the filesystem root. The
// operation is atomic (written to a temp file and renamed). Returns an
// error if the resolved path would be outside the configured Root. Parent
// directories MUST exist before calling WriteFile; this method will NOT
// create them.
func (fs osFileSystem) WriteFile(filename string, data []byte) error {
	if filepath.IsAbs(filename) {
		return fmt.Errorf("filename must be relative to root: %s", filename)
	}
	fullPath := fs.fullPath(filename)
	if !fs.isUnderRoot(fullPath) {
		return fmt.Errorf("path escapes root: %s", filename)
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

// ReadFile reads the content of filename from the filesystem root. It fails
// if the resolved path would escape the configured Root.
func (fs osFileSystem) ReadFile(filename string) ([]byte, error) {
	if filepath.IsAbs(filename) {
		return nil, fmt.Errorf("filename must be relative to root: %s", filename)
	}
	fullPath := fs.fullPath(filename)
	if !fs.isUnderRoot(fullPath) {
		return nil, fmt.Errorf("path escapes root: %s", filename)
	}
	return os.ReadFile(fullPath) // #nosec G304 -- fullPath is validated via isUnderRoot
}

// ReadDir lists all entries directly under path (relative to the
// filesystem root) and returns their base names. The operation is rejected
// if the resolved path would escape the configured Root.
func (fs osFileSystem) ReadDir(path string) ([]string, error) {
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("path must be relative to root: %s", path)
	}
	fullPath := fs.fullPath(path)
	if !fs.isUnderRoot(fullPath) {
		return nil, fmt.Errorf("path escapes root: %s", path)
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

// Remove removes the named path from the filesystem root. When recursive is
// true the path and all children are removed; when false only the named
// path is removed and the call will fail for non-empty directories.
//
// The operation is rejected if the path would escape the configured Root.
// If the path does not exist the method returns nil.
func (fs osFileSystem) Remove(path string, recursive bool) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative to root: %s", path)
	}
	full := fs.fullPath(path)
	if !fs.isUnderRoot(full) {
		return fmt.Errorf("path escapes root: %s", path)
	}

	var err error
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

func (fs osFileSystem) fullPath(p string) string {
	return filepath.Clean(filepath.Join(fs.Root, p))
}

// isUnderRoot reports whether the cleaned path p is located inside the
// filesystem root. It uses filepath.Rel to compute the relative path from
// the root to p; if the relative path begins with ".." the path is outside
// the root.
func (fs osFileSystem) isUnderRoot(p string) bool {
	fullPath := filepath.Clean(p)

	rel, err := filepath.Rel(fs.Root, fullPath)
	if err != nil {
		return false
	}

	if rel == "." {
		return true
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return true
}
