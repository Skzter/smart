package repository

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rootDir string = "testTmp"

func TestNewOSFileSystem(t *testing.T) {
	tests := []struct {
		name            string
		root            string
		expectError     bool
		expectNilResult bool
	}{
		{
			name:            "empty root",
			root:            "",
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "valid root",
			root:            "temp",
			expectError:     false,
			expectNilResult: false,
		},
	}

	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal("setup failed")
	}

	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs, err := NewOSFileSystem(test.root)

			if test.expectError {
				if err == nil {
					t.Errorf("NewOSFileSystem() expected error but got none")
				}
				if !test.expectNilResult && fs != nil {
					t.Errorf("NewOSFileSystem() expected nil result on error, got: %+v", fs)
				}
				return
			} else {
				if err != nil {
					t.Errorf("NewOSFileSystem() unexpected error: %v", err)
				}
				if fs == nil {
					t.Errorf("NewOSFileSystem() expected non-nil result on success")
				}
			}
		})
	}
}

func TestNewLogFileSystem(t *testing.T) {
	tests := []struct {
		name        string
		root        string
		expectError bool
	}{
		{
			name:        "empty root",
			root:        "",
			expectError: true,
		},
		{
			name:        "valid root",
			root:        "temp",
			expectError: false,
		},
	}

	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal("setup failed")
	}

	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewLogFileSystem(tc.root)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, fs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fs)
			}
		})
	}
}

func TestMkdirAll(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errContains string
		expectExist bool
	}{
		{
			name:        "create nested relative path",
			path:        "a/b/c",
			expectError: false,
			expectExist: true,
		},
		{
			name:        "empty path (root) is allowed",
			path:        "",
			expectError: false,
			expectExist: true,
		},
		{
			name:        "absolute path rejected",
			path:        "/abs/path",
			expectError: true,
			errContains: "relative to root",
			expectExist: false,
		},
		{
			name:        "path escapes root with .. rejected",
			path:        "../outside",
			expectError: true,
			errContains: "escapes root",
			expectExist: false,
		},
	}

	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal("setup failed")
	}

	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}

	fs, err := NewOSFileSystem(rootDir)
	if err != nil {
		t.Fatal("setup failed")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fs.MkdirAll(test.path)
			if test.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.expectExist {
				full := filepath.Join(rootDir, test.path)
				info, statErr := os.Stat(full)
				if statErr != nil {
					t.Fatalf("expected path to exist (%s): stat error: %v", full, statErr)
				}
				if !info.IsDir() {
					t.Fatalf("expected %s to be a directory", full)
				}
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name             string
		relativeFilePath string
		data             []byte
		expectError      bool
	}{
		{
			name:             "relativeFilePath is empty",
			relativeFilePath: "",
			data:             nil,
			expectError:      true,
		},
		{
			name:             "relativeFilePath is an absolute path",
			relativeFilePath: "/abs/path/file.txt",
			data:             nil,
			expectError:      true,
		},
		{
			name:             "relativeFilePath is valid",
			relativeFilePath: "validFilename.txt",
			data:             []byte("Test data"),
			expectError:      false,
		},
		{
			name:             "relativeFilePath escapes root",
			relativeFilePath: "../validFilename.txt",
			data:             []byte("Test data"),
			expectError:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir to temp dir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatal("setup failed")
			}

			resultErr := fs.WriteFile(test.relativeFilePath, test.data)

			if test.expectError {
				if resultErr == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			fullPath := filepath.Join(tmp, rootDir, test.relativeFilePath)
			checkedFullPath, err := filepath.Abs(fullPath)
			if err != nil {
				t.Fatalf("setup failed")
			}
			if r, err := filepath.EvalSymlinks(checkedFullPath); err == nil {
				checkedFullPath = r
			}

			content, readErr := os.ReadFile(checkedFullPath)
			if readErr != nil {
				t.Fatalf("expected file %q to exist: %v", fullPath, readErr)
			}

			if string(content) != string(test.data) {
				t.Fatalf("file content mismatch:\nwant: %q\ngot:  %q", string(test.data), string(content))
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name             string
		relativeFilePath string
		expectError      bool
		initialData      []byte
		expectedData     []byte
	}{
		{
			name:             "relativeFilePath is empty",
			relativeFilePath: "",
			expectError:      true,
		},
		{
			name:             "relativeFilePath is an absolute path",
			relativeFilePath: "/abs/path",
			expectError:      true,
		},
		{
			name:             "relativeFilePath is valid",
			relativeFilePath: "validFilename",
			initialData:      []byte("Test data"),
			expectError:      false,
			expectedData:     []byte("Test data"),
		},
		{
			name:             "relativeFilePath escapes root",
			relativeFilePath: "../validFilename",
			expectError:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir to temp dir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatalf("NewOSFileSystem failed: %v", err)
			}

			if test.initialData != nil {
				fullPath := filepath.Join(tmp, rootDir, test.relativeFilePath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
					t.Fatalf("mkdir failed: %v", err)
				}
				if err := os.WriteFile(fullPath, test.initialData, 0o600); err != nil {
					t.Fatalf("write file failed: %v", err)
				}
			}

			data, resultErr := fs.ReadFile(test.relativeFilePath)

			if test.expectError {
				if resultErr == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}

			if resultErr != nil {
				t.Fatalf("unexpected error: %v", resultErr)
			}

			if string(data) != string(test.expectedData) {
				t.Fatalf("file content mismatch:\nwant: %q\ngot:  %q", string(test.expectedData), string(data))
			}
		})
	}
}

func TestReadDir(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		initialStructure []string // paths relative to rootDir; trailing '/' marks a directory
		expectError      bool
		expectedData     []string
	}{
		{
			name:        "path is absolute",
			path:        "/abs/path",
			expectError: true,
		},
		{
			name:             "list files in root",
			path:             "",
			initialStructure: []string{"file1.txt", "file2.txt", "dir1/"},
			expectError:      false,
			expectedData:     []string{"file1.txt", "file2.txt", "dir1"},
		},
		{
			name:             "list files in subdir",
			path:             "sub",
			initialStructure: []string{"sub/fileA.txt", "sub/fileB.txt", "sub/nested/"},
			expectError:      false,
			expectedData:     []string{"fileA.txt", "fileB.txt", "nested"},
		},
		{
			name:             "path escapes root",
			path:             "../outside",
			initialStructure: nil,
			expectError:      true,
			expectedData:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir to temp dir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatalf("NewOSFileSystem failed: %v", err)
			}

			if test.initialStructure != nil {
				for _, rel := range test.initialStructure {
					fullPath := filepath.Join(tmp, rootDir, rel)
					if strings.HasSuffix(rel, "/") {
						if err := os.MkdirAll(fullPath, 0o750); err != nil {
							t.Fatalf("mkdir failed for %s: %v", fullPath, err)
						}
					} else {
						if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
							t.Fatalf("mkdir failed for %s: %v", fullPath, err)
						}
						if err := os.WriteFile(fullPath, []byte("x"), 0o600); err != nil {
							t.Fatalf("write file failed for %s: %v", fullPath, err)
						}
					}
				}
			}

			names, err := fs.ReadDir(test.path)

			if test.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			sort.Strings(names)
			sort.Strings(test.expectedData)
			if len(names) != len(test.expectedData) {
				t.Fatalf("unexpected entries: want=%v got=%v", test.expectedData, names)
			}
			for i := range names {
				if names[i] != test.expectedData[i] {
					t.Fatalf("unexpected entries: want=%v got=%v", test.expectedData, names)
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := removeTests()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir to temp dir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatalf("NewOSFileSystem failed: %v", err)
			}

			if test.initialStructure != nil {
				for _, rel := range test.initialStructure {
					fullPath := filepath.Join(tmp, rootDir, rel)
					if strings.HasSuffix(rel, "/") {
						if err := os.MkdirAll(fullPath, 0o750); err != nil {
							t.Fatalf("mkdir failed for %s: %v", fullPath, err)
						}
					} else {
						if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
							t.Fatalf("mkdir failed for %s: %v", fullPath, err)
						}
						if err := os.WriteFile(fullPath, []byte("x"), 0o600); err != nil {
							t.Fatalf("write file failed for %s: %v", fullPath, err)
						}
					}
				}
			}

			err = fs.Remove(test.path, test.recursive)

			if test.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.expectExistsAfter {
				full := filepath.Join(tmp, rootDir, test.path)
				if _, statErr := os.Stat(full); statErr != nil {
					t.Fatalf("expected %s to exist, stat error: %v", full, statErr)
				}
			} else {
				full := filepath.Join(tmp, rootDir, test.path)
				if _, statErr := os.Stat(full); statErr == nil {
					t.Fatalf("expected %s to be removed, but it exists", full)
				} else if !os.IsNotExist(statErr) {
					t.Fatalf("unexpected stat error for %s: %v", full, statErr)
				}
			}
		})
	}
}

func removeTests() []struct {
	name              string
	path              string
	recursive         bool
	initialStructure  []string
	expectError       bool
	expectExistsAfter bool
} {
	return []struct {
		name              string
		path              string
		recursive         bool
		initialStructure  []string
		expectError       bool
		expectExistsAfter bool
	}{
		{
			name:              "remove non-existent",
			path:              "nope",
			recursive:         false,
			initialStructure:  nil,
			expectError:       false,
			expectExistsAfter: false,
		},
		{
			name:              "remove file",
			path:              "file.txt",
			recursive:         false,
			initialStructure:  []string{"file.txt"},
			expectError:       false,
			expectExistsAfter: false,
		},
		{
			name:              "remove empty dir non-recursive",
			path:              "emptydir",
			recursive:         false,
			initialStructure:  []string{"emptydir/"},
			expectError:       false,
			expectExistsAfter: false,
		},
		{
			name:              "remove non-empty dir non-recursive",
			path:              "dir",
			recursive:         false,
			initialStructure:  []string{"dir/", "dir/child.txt"},
			expectError:       true,
			expectExistsAfter: true,
		},
		{
			name:              "remove non-empty dir recursive",
			path:              "dir",
			recursive:         true,
			initialStructure:  []string{"dir/", "dir/child.txt"},
			expectError:       false,
			expectExistsAfter: false,
		},
		{
			name:        "absolute path rejected",
			path:        filepath.Join(string(os.PathSeparator), "abs", "path"),
			recursive:   false,
			expectError: true,
		},
		{
			name:        "path escapes root",
			path:        "../outside",
			recursive:   false,
			expectError: true,
		},
	}
}

func TestGetFileStats(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		setupFile   bool
		expectError bool
	}{
		{
			name:        "absolute path",
			path:        "/abs/path",
			expectError: true,
		},
		{
			name:        "file stats",
			path:        "file.txt",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "nested file",
			path:        "sub/file.txt",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "path escapes root",
			path:        "../outside",
			expectError: true,
		},
		{
			name:        "non-existent file",
			path:        "missing.txt",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatalf("NewOSFileSystem: %v", err)
			}

			if test.setupFile {
				fullPath := filepath.Join(tmp, rootDir, test.path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte("content"), 0o600); err != nil {
					t.Fatalf("write: %v", err)
				}
			}

			info, err := fs.GetFileStats(test.path)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if info.Size() == 0 {
				t.Fatal("expected file size > 0")
			}
		})
	}
}

func TestGetValidatedPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		setupFile   bool
	}{
		{
			name:        "absolute path rejected",
			path:        "/abs/path",
			expectError: true,
		},
		{
			name:        "path escapes root",
			path:        "../outside",
			expectError: true,
		},
		{
			name:        "valid relative path",
			path:        "sub/file.txt",
			expectError: false,
			setupFile:   true,
		},
		{
			name:        "valid root-level path",
			path:        "file.txt",
			expectError: false,
			setupFile:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatal("setup failed")
			}

			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Fatalf("restore working dir: %v", err)
				}
			}()

			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("chdir: %v", err)
			}

			fs, err := NewOSFileSystem(rootDir)
			if err != nil {
				t.Fatalf("NewOSFileSystem: %v", err)
			}

			if test.setupFile {
				fullPath := filepath.Join(tmp, rootDir, test.path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte("content"), 0o600); err != nil {
					t.Fatalf("write: %v", err)
				}
			}

			validatedPath, err := fs.GetValidatedPath(test.path)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if validatedPath == "" {
				t.Fatal("expected non-empty validated path")
			}

			expectedPath := filepath.Clean(filepath.Join(rootDir, test.path))
			if validatedPath != expectedPath {
				t.Fatalf("path mismatch:\nwant: %q\ngot:  %q", expectedPath, validatedPath)
			}
		})
	}
}

func TestIsUnderRoot(t *testing.T) {
	tmp := t.TempDir()
	tmpEvaluated, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatalf("setup failed")
	}
	fs := &osFileSystem{Root: tmpEvaluated}

	subdir := filepath.Join(tmpEvaluated, "subdir")
	if err := os.Mkdir(subdir, DefaultDirPerm); err != nil {
		t.Fatalf("setup: create subdir: %v", err)
	}

	aDir := filepath.Join(tmpEvaluated, "a")
	if err := os.Mkdir(aDir, DefaultDirPerm); err != nil {
		t.Fatalf("setup: create a dir: %v", err)
	}
	bDir := filepath.Join(tmpEvaluated, "b")
	if err := os.Mkdir(bDir, DefaultDirPerm); err != nil {
		t.Fatalf("setup: create b dir: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedResult bool
	}{
		{
			name:           "path is exactly root",
			path:           tmpEvaluated,
			expectedResult: true,
		},
		{
			name:           "path is inside root",
			path:           subdir,
			expectedResult: true,
		},
		{
			name:           "path escapes root",
			path:           filepath.Join(filepath.Dir(tmpEvaluated), "outside"),
			expectedResult: false,
		},
		{
			name:           "path contains .. but stays inside root",
			path:           filepath.Join(tmpEvaluated, "a", "..", "b"),
			expectedResult: true,
		},
		{
			name:           "path far outside root (../..)",
			path:           filepath.Join(tmpEvaluated, "..", "..", "escape"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := fs.isUnderRoot(test.path)

			if result != test.expectedResult {
				t.Fatalf("expected %t but got %t \npath was: %s", test.expectedResult, result, test.path)
			}
		})
	}
}
