package assert

import (
	"github.com/pkg/errors"
)

func NumberBetween(value interface{}, minValue interface{}, maxValue interface{}) error {
	if err := checkNumber(value, &minValue, &maxValue, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func NumberGreaterThan(value interface{}, minValue interface{}) error {
	if err := checkNumber(value, &minValue, nil, true); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func NumberGreaterOrEqualThan(value interface{}, minValue interface{}) error {
	if err := checkNumber(value, &minValue, nil, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func NumberLessThan(value interface{}, maxValue interface{}) error {
	if err := checkNumber(value, nil, &maxValue, true); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func NumberLessOrEqualThan(value interface{}, maxValue interface{}) error {
	if err := checkNumber(value, nil, &maxValue, false); err != nil {
		return errors.Wrap(err, "Assert failed")
	}

	return nil
}

func Id(value interface{}) error {
	return NumberGreaterThan(value, 0)
}
