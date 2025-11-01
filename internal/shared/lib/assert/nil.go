package assert

import (
	"fmt"
	"reflect"
)

// NotNil checks that all provided values are not nil.
// Returns an error if any value is nil, otherwise returns nil.
func NotNil(values ...interface{}) error {
	for i, value := range values {
		if value == nil {
			return fmt.Errorf("assert failed: given value at index %d is nil", i)
		}

		switch reflect.TypeOf(value).Kind() {
		case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Func:
			if reflect.ValueOf(value).IsNil() {
				return fmt.Errorf("assert failed: given value at index %d is nil", i)
			}
		}
	}

	return nil
}
