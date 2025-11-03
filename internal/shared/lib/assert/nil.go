package assert

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// NotNil checks that all provided values are not nil.
// Returns an error if any value is nil, otherwise returns nil.
func NotNil(values ...interface{}) error {
	for i, value := range values {
		if value == nil {
			return errors.WithStack(fmt.Errorf("assert failed: given value at index %d is nil", i))
		}

		switch reflect.TypeOf(value).Kind() {
		case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Func:
			if reflect.ValueOf(value).IsNil() {
				return errors.WithStack(fmt.Errorf("assert failed: given value at index %d is nil", i))
			}
		}
	}

	return nil
}
