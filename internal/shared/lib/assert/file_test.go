//nolint:gosec,staticcheck
package assert

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
