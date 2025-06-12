//nolint:dupl
package assert

import (
	"github.com/pkg/errors"
)

func MapLengthBetween(value interface{}, minValue int, maxValue int) error {
	if err := checkMapLength(value, &minValue, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func MapLengthGreaterThan(value interface{}, minValue int) error {
	minValue++

	if err := checkMapLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func MapLengthGreaterOrEqualThan(value interface{}, minValue int) error {
	if err := checkMapLength(value, &minValue, nil); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func MapLengthLessThan(value interface{}, maxValue int) error {
	maxValue--

	if err := checkMapLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func MapLengthLessOrEqualThan(value interface{}, maxValue int) error {
	if err := checkMapLength(value, nil, &maxValue); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}
