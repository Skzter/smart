//nolint:gosec,staticcheck
package assert

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileExists tests the FileExists function.
// It checks that the function returns nil for an existing file
// and returns an error for a non-existing file.
func TestFileExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	existingFile, err := os.Create(tmpDir + "/existingFile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, FileExists(existingFile.Name()))
	assert.NotNil(t, FileExists(tmpDir+"/fileThatDoesNotExist"))
}
