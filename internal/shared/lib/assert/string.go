package assert

import (
	"github.com/pkg/errors"
)

func ByteLength(value []byte, minLength int, maxLength int) error {
	if minLength != -1 && maxLength != -1 && minLength > maxLength {
		return errors.Errorf("Assert misuse: minLength should be smaller than maxLength (minLength: %d, maxLength: %d)", minLength, maxLength)
	}

	lenBytes := len(value)
	if minLength != -1 && lenBytes < minLength {
		return errors.Errorf("Assert failed: given value '%s' is too short (%d < %d)", value, lenBytes, minLength)
	}

	if maxLength != -1 && lenBytes > maxLength {
		return errors.Errorf("Assert failed: given value '%s' is too long (%d > %d)", value, lenBytes, minLength)
	}

	return nil
}

// StringNotEmpty checks if given string is not empty.
func StringNotEmpty(value string) error {
	return StringLength(value, 1, -1, nil)
}

func StringLength(value string, minLength int, maxLength int, allowedValues *[]string) error {
	if minLength != -1 && maxLength != -1 && minLength > maxLength {
		return errors.Errorf("Assert misuse: minLength should be smaller than maxLength (minLength: %d, maxLength: %d)", minLength, maxLength)
	}

	lenString := len(value)
	if minLength != -1 && lenString < minLength {
		return errors.Errorf("Assert failed: given value '%s' is too short (%d < %d)", value, lenString, minLength)
	}

	if maxLength != -1 && lenString > maxLength {
		return errors.Errorf("Assert failed: given value '%s' is too long (%d > %d)", value, lenString, minLength)
	}

	if allowedValues != nil {
		found := false

		for _, allowedValue := range *allowedValues {
			if allowedValue == value {
				found = true
				break
			}
		}

		if !found {
			return errors.Errorf("Assert failed: given value '%s' is not in allowed values (%s)", value, *allowedValues)
		}
	}

	return nil
}
