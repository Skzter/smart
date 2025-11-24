package assert

import (
	"github.com/pkg/errors"
)

// ByteLength checks if the length of the byte slice is within the given range.
// Returns an error if the length is outside the range.
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
// Returns an error if the string is empty.
func StringNotEmpty(value string) error {
	return StringLength(value, 1, -1, nil)
}

// StringsNotEmpty checks if all of the given strings are not empty.
// Returns an error if any string is empty.
func StringsNotEmpty(values ...string) error {
	for i, val := range values {
		if StringNotEmpty(val) != nil {
			return errors.Errorf("Assert failed: given string at index %d is empty", i)
		}
	}
	return nil
}

// StringLength checks if the string length is within the given range and optionally if it is in the allowed values.
// Returns an error if the string does not meet the requirements.
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
