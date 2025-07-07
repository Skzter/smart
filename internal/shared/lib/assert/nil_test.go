package assert

import (
	"fmt"
	"testing"
)

type testData struct {
	value int
}

// TestNotNil tests the NotNil function.
// It checks that NotNil returns an error if any value is nil and nil otherwise.
func TestNotNil(t *testing.T) {
	testDataNotNil := testData{value: 5}
	testDataNotNilPointer := &testDataNotNil
	var testDataNilPointer *testData
	var testFuncNil func()
	testFuncNotNil := func() {}

	tests := []struct {
		name     string
		value    []interface{}
		error    bool
		nilIndex int
	}{
		{
			name:  "int",
			value: []interface{}{10},
			error: false,
		},
		{
			name:  "string",
			value: []interface{}{"hello"},
			error: false,
		},
		{
			name:  "struct",
			value: []interface{}{testDataNotNil},
			error: false,
		},
		{
			name:  "not nil pointer",
			value: []interface{}{testDataNotNilPointer},
			error: false,
		},
		{
			name:  "nil literal",
			value: []interface{}{nil},
			error: true,
		},
		{
			name:  "nil pointer",
			value: []interface{}{testDataNilPointer},
			error: true,
		},
		{
			name:  "nil function",
			value: []interface{}{testFuncNil},
			error: true,
		},
		{
			name:  "not nil function",
			value: []interface{}{testFuncNotNil},
			error: false,
		},
		{
			name:  "slice of not nil elements",
			value: []interface{}{testFuncNotNil, testDataNotNilPointer, "abc", 10},
			error: false,
		},
		{
			name:     "slice of not nil elements except one nil pointer @ index 1",
			value:    []interface{}{testFuncNotNil, testDataNilPointer, "abc", 10},
			error:    true,
			nilIndex: 1,
		},
		{
			name:     "slice of only nil elements",
			value:    []interface{}{testFuncNil, testDataNilPointer, nil},
			error:    true,
			nilIndex: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NotNil(test.value...)
			switch {
			case test.error && err == nil:
				t.Errorf("Got nil, expected error (value: %d)", test.value)
			case test.error && err != nil: // check whether the error message is correct
				msg := err.Error()
				expectedMsg := fmt.Sprintf("assert failed: given value at index %d is nil", test.nilIndex)
				if expectedMsg != msg {
					t.Errorf("Got error message '%s', expected error message '%s'", expectedMsg, msg)
				}
			case !test.error && err != nil:
				t.Errorf("Got error '%s', expected nil (value: %d)", err, test.value)
			}
		})
	}
}
