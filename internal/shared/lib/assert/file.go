package assert

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// FileExists checks if a file exists at the given path.
// It returns an error if the file does not exist.
func FileExists(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("file %s doesnt exist: %s", filename, err))
	}

	return nil
}
