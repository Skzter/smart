//nolint:dupl
package assert

import (
	"github.com/pkg/errors"
)

func ArrayLengthBetween(value interface{}, minValue int, maxValue int) error {
	if err := checkArrayLength(value, &minValue, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func ArrayLengthGreaterThan(value interface{}, minValue int) error {
	minValue++

	if err := checkArrayLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func ArrayLengthGreaterOrEqualThan(value interface{}, minValue int) error {
	if err := checkArrayLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func ArrayLengthLessThan(value interface{}, maxValue int) error {
	maxValue--

	if err := checkArrayLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func ArrayLengthLessOrEqualThan(value interface{}, maxValue int) error {
	if err := checkArrayLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}
