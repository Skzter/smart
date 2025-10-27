package assert

import (
	"fmt"
	"reflect"
)

// NotNilError is custom error for notnil assertions
type NotNilError struct {
	Message string
}

func (e *NotNilError) Error() string {
	return e.Message
}

// NotNil checks that all provided values are not nil.
// Returns an error if any value is nil, otherwise returns nil.
func NotNil(values ...interface{}) error {
	for i, value := range values {
		if value == nil {
			return &NotNilError{
				Message: fmt.Sprintf("assert failed: given value at index %d is nil", i),
			}
		}

		switch reflect.TypeOf(value).Kind() {
		case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Func:
			if reflect.ValueOf(value).IsNil() {
				return &NotNilError{
					Message: fmt.Sprintf("assert failed: given value at index %d is nil", i),
				}
			}
		}
	}

	return nil
}
