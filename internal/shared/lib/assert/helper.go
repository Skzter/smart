//nolint:gosec
package assert

import (
	"reflect"

	"github.com/pkg/errors"
)

func convertToInt64(value interface{}) (int64, bool) {
	if v, ok := value.(int); ok {
		return int64(v), true
	}
	if v, ok := value.(int64); ok {
		return v, true
	}
	if v, ok := value.(int8); ok {
		return int64(v), true
	}
	if v, ok := value.(int16); ok {
		return int64(v), true
	}
	if v, ok := value.(int32); ok {
		return int64(v), true
	}
	return 0, false
}

func convertToUint64(value interface{}) (uint64, bool) {
	if v, ok := value.(uint8); ok {
		return uint64(v), true
	}
	if v, ok := value.(int); ok {
		return uint64(v), true //lint:ignore G115 is only a helper
	}
	if v, ok := value.(uint32); ok {
		return uint64(v), true
	}
	if v, ok := value.(uint); ok {
		return uint64(v), true
	}
	if v, ok := value.(uint64); ok {
		return v, true
	}
	if v, ok := value.(uint16); ok {
		return uint64(v), true
	}
	if v, ok := value.(int64); ok {
		return uint64(v), true //lint:ignore G115 is only a helper
	}
	if v, ok := value.(int8); ok {
		return uint64(v), true //lint:ignore G115 is only a helper
	}
	if v, ok := value.(int16); ok {
		return uint64(v), true
	}
	if v, ok := value.(int32); ok {
		return uint64(v), true
	}
	return 0, false
}

func convertToFloat64(value interface{}) (float64, bool) {
	if v, ok := value.(float64); ok {
		return v, true
	}
	if v, ok := value.(float32); ok {
		return float64(v), true
	}
	return 0, false
}

//nolint:funlen,gocyclo
func checkNumber(value interface{}, minValue *interface{}, maxValue *interface{}, closedBoundaries bool) error {
	switch (value).(type) {
	case int, int8, int16, int32, int64:
		vInt, _ := convertToInt64(value)
		if minValue != nil {
			min, isInt := convertToInt64(*minValue)
			if !isInt {
				return errors.Errorf("Min Value is no signed integer type (%T)", *minValue)
			}
			if closedBoundaries {
				if vInt <= min {
					return errors.Errorf("Given value %d is too small (%d <= %d)", vInt, vInt, min)
				}
			} else {
				if vInt < min {
					return errors.Errorf("Given value %d is too small (%d < %d)", vInt, vInt, min)
				}
			}
		}

		if maxValue != nil {
			max, isInt := convertToInt64(*maxValue)
			if !isInt {
				return errors.Errorf("Max Value is no signed integer type (%T)", *maxValue)
			}
			if closedBoundaries {
				if vInt >= max {
					return errors.Errorf("Given value %d is too big (%d >= %d)", vInt, vInt, max)
				}
			} else {
				if vInt > max {
					return errors.Errorf("Given value %d is too big (%d > %d)", vInt, vInt, max)
				}
			}
		}
	case uint, uint8, uint16, uint32, uint64:
		vUInt, _ := convertToUint64(value)
		if minValue != nil {
			min, isUint := convertToUint64(*minValue)
			if !isUint {
				return errors.Errorf("Min Value is no integer type (%T)", *minValue)
			}
			if closedBoundaries {
				if vUInt <= min {
					return errors.Errorf("Given value %d is too small (%d <= %d)", vUInt, vUInt, min)
				}
			} else {
				if vUInt < min {
					return errors.Errorf("Given value %d is too small (%d < %d)", vUInt, vUInt, min)
				}
			}
		}

		if maxValue != nil {
			max, isUint := convertToUint64(*maxValue)
			if !isUint {
				return errors.Errorf("Max Value is no integer type (%T)", *maxValue)
			}
			if closedBoundaries {
				if vUInt >= max {
					return errors.Errorf("Given value %d is too big (%d >= %d)", vUInt, vUInt, max)
				}
			} else {
				if vUInt > max {
					return errors.Errorf("Given value %d is too big (%d > %d)", vUInt, vUInt, max)
				}
			}
		}
	case float32, float64:
		vFloat, _ := convertToFloat64(value)
		if minValue != nil {
			min, isFloat := convertToFloat64(*minValue)
			if !isFloat {
				return errors.Errorf("Min Value is no float type (%T)", *minValue)
			}
			if closedBoundaries {
				if vFloat <= min {
					return errors.Errorf("Given value %f is too small (%f <= %f)", vFloat, vFloat, min)
				}
			} else {
				if vFloat < min {
					return errors.Errorf("Given value %f is too small (%f < %f)", vFloat, vFloat, min)
				}
			}
		}

		if maxValue != nil {
			max, isFloat := convertToFloat64(*maxValue)
			if !isFloat {
				return errors.Errorf("Max Value is no float type (%T)", *maxValue)
			}
			if closedBoundaries {
				if vFloat >= max {
					return errors.Errorf("Given value %f is too big (%f >= %f)", vFloat, vFloat, max)
				}
			} else {
				if vFloat > max {
					return errors.Errorf("Given value %f is too big (%f > %f)", vFloat, vFloat, max)
				}
			}
		}
	default:
		return errors.Errorf("Value is no number type (%T)", value)
	}
	return nil
}

func arrayLength(value interface{}) (int, error) {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
		return reflect.ValueOf(value).Len(), nil
	case reflect.Slice:
		return reflect.ValueOf(value).Len(), nil
	}

	return 0, errors.Errorf("Value must be array or slice ('%T')", value)
}

func checkArrayLength(value interface{}, minValue *int, maxValue *int) error {
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return errors.Errorf("minValue should be smaller than maxValue (minValue: %d, maxValue: %d)", *minValue, *maxValue)
	}

	arrayLen, err := arrayLength(value)
	if err != nil {
		return err
	}

	if minValue != nil {
		if arrayLen < *minValue {
			return errors.Errorf("Length %d is too small (%d < %d)", arrayLen, arrayLen, *minValue)
		}
	}

	if maxValue != nil {
		if arrayLen > *maxValue {
			return errors.Errorf("Length %d is too big (%d > %d)", arrayLen, arrayLen, *maxValue)
		}
	}

	return nil
}

func mapLength(value interface{}) (int, error) {
	if reflect.TypeOf(value).Kind() == reflect.Map {
		return reflect.ValueOf(value).Len(), nil
	}

	return 0, errors.Errorf("Value must be map ('%T')", value)
}

func checkMapLength(value interface{}, minValue *int, maxValue *int) error {
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return errors.Errorf("minValue should be smaller than maxValue (minValue: %d, maxValue: %d)", *minValue, *maxValue)
	}

	mapLen, err := mapLength(value)
	if err != nil {
		return err
	}

	if minValue != nil {
		if mapLen < *minValue {
			return errors.Errorf("Length %d is too small (%d < %d)", mapLen, mapLen, *minValue)
		}
	}

	if maxValue != nil {
		if mapLen > *maxValue {
			return errors.Errorf("Length %d is too big (%d > %d)", mapLen, mapLen, *maxValue)
		}
	}

	return nil
}
