//nolint:dupl
package assert

import (
	"github.com/pkg/errors"
)

// MapLengthBetween checks if the length of the map is between minValue and maxValue (inclusive).
func MapLengthBetween(value interface{}, minValue int, maxValue int) error {
	if err := checkMapLength(value, &minValue, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// MapLengthGreaterThan checks if the length of the map is greater than minValue.
func MapLengthGreaterThan(value interface{}, minValue int) error {
	minValue++

	if err := checkMapLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// MapLengthGreaterOrEqualThan checks if the length of the map is greater than or equal to minValue.
func MapLengthGreaterOrEqualThan(value interface{}, minValue int) error {
	if err := checkMapLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// MapLengthLessThan checks if the length of the map is less than maxValue.
func MapLengthLessThan(value interface{}, maxValue int) error {
	maxValue--

	if err := checkMapLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// MapLengthLessOrEqualThan checks if the length of the map is less than or equal to maxValue.
func MapLengthLessOrEqualThan(value interface{}, maxValue int) error {
	if err := checkMapLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}
