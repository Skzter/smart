package assert

import (
	"github.com/pkg/errors"
)

// NumberBetween checks if value is between minValue and maxValue (inclusive).
// Returns an error if the value is not within the range.
func NumberBetween(value interface{}, minValue interface{}, maxValue interface{}) error {
	if err := checkNumber(value, &minValue, &maxValue, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// NumberGreaterThan checks if value is strictly greater than minValue.
// Returns an error if the value is not greater.
func NumberGreaterThan(value interface{}, minValue interface{}) error {
	if err := checkNumber(value, &minValue, nil, true); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// NumberGreaterOrEqualThan checks if value is greater than or equal to minValue.
// Returns an error if the value is less.
func NumberGreaterOrEqualThan(value interface{}, minValue interface{}) error {
	if err := checkNumber(value, &minValue, nil, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// NumberLessThan checks if value is strictly less than maxValue.
// Returns an error if the value is not less.
func NumberLessThan(value interface{}, maxValue interface{}) error {
	if err := checkNumber(value, nil, &maxValue, true); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// NumberLessOrEqualThan checks if value is less than or equal to maxValue.
// Returns an error if the value is greater.
func NumberLessOrEqualThan(value interface{}, maxValue interface{}) error {
	if err := checkNumber(value, nil, &maxValue, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

// Id checks if value is greater than 0.
// Returns an error if the value is 0 or less.
func Id(value interface{}) error {
	return NumberGreaterThan(value, 0)
}
