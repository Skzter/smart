package assert

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Assert that a file exists
func FileExists(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("file %s doesnt exist: %s", filename, err))
	}

	return nil
}
