//nolint:dupl
package assert

import (
	"github.com/pkg/errors"
)

// ArrayLengthBetween checks if the length of the array is between minValue and maxValue (inclusive).
func ArrayLengthBetween(value interface{}, minValue int, maxValue int) error {
	if err := checkArrayLength(value, &minValue, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}
	return nil
}

// ArrayLengthGreaterThan checks if the length of the array is greater than minValue.
func ArrayLengthGreaterThan(value interface{}, minValue int) error {
	minValue++

	if err := checkArrayLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}
	return nil
}

// ArrayLengthGreaterOrEqualThan checks if the length of the array is greater than or equal to minValue.
func ArrayLengthGreaterOrEqualThan(value interface{}, minValue int) error {
	if err := checkArrayLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}
	return nil
}

// ArrayLengthLessThan checks if the length of the array is less than maxValue.
func ArrayLengthLessThan(value interface{}, maxValue int) error {
	maxValue--

	if err := checkArrayLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}
	return nil
}

// ArrayLengthLessOrEqualThan checks if the length of the array is less than or equal to maxValue.
func ArrayLengthLessOrEqualThan(value interface{}, maxValue int) error {
	if err := checkArrayLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}
	return nil
}
