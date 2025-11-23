package repository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
