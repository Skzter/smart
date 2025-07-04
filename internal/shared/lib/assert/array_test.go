package assert

import (
	"testing"
)

// TestArrayLengthBetween checks if ArrayLengthBetween correctly validates array length within a range.
func TestArrayLengthBetween(t *testing.T) {
	tests := []struct {
		id       int
		value    [3]int
		minValue int
		maxValue int
		error    bool
	}{
		{1, [3]int{1, 2, 3}, 1, 5, false},
		{2, [3]int{1, 2, 3}, 1, 3, false},
		{3, [3]int{1, 2, 3}, 3, 5, false},
		{4, [3]int{1, 2, 3}, 1, 2, true},
		{5, [3]int{1, 2, 3}, 4, 5, true},
		{6, [3]int{1, 2, 3}, 2, 1, true},
	}

	for _, test := range tests {
		err := ArrayLengthBetween(test.value, test.minValue, test.maxValue)

		// Check if the error result matches the expected outcome.
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}
}

// TestArrayLengthGreaterThan checks if ArrayLengthGreaterThan correctly validates array length is greater than minValue.
func TestArrayLengthGreaterThan(t *testing.T) {
	tests := []struct {
		id       int
		value    [3]int
		minValue int
		error    bool
	}{
		{1, [3]int{1, 2, 3}, 3, true},
		{2, [3]int{1, 2, 3}, 1, false},
		{4, [3]int{1, 2, 3}, 5, true},
	}

	for _, test := range tests {
		err := ArrayLengthGreaterThan(test.value, test.minValue)

		// Check if the error result matches the expected outcome.
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

// TestArrayLengthGreaterOrEqualThan checks if ArrayLengthGreaterOrEqualThan validates array length is greater or equal to minValue.
func TestArrayLengthGreaterOrEqualThan(t *testing.T) {
	tests := []struct {
		id       int
		value    [3]int
		minValue int
		error    bool
	}{
		{1, [3]int{1, 2, 3}, 3, false},
		{2, [3]int{1, 2, 3}, 1, false},
		{4, [3]int{1, 2, 3}, 5, true},
	}

	for _, test := range tests {
		err := ArrayLengthGreaterOrEqualThan(test.value, test.minValue)

		// Check if the error result matches the expected outcome.
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

// TestArrayLengthLessThan checks if ArrayLengthLessThan validates array length is less than maxValue.
func TestArrayLengthLessThan(t *testing.T) {
	tests := []struct {
		id       int
		value    [3]int
		maxValue int
		error    bool
	}{
		{1, [3]int{1, 2, 3}, 3, true},
		{2, [3]int{1, 2, 3}, 1, true},
		{4, [3]int{1, 2, 3}, 5, false},
	}

	for _, test := range tests {
		err := ArrayLengthLessThan(test.value, test.maxValue)

		// Check if the error result matches the expected outcome.
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}

// TestArrayLengthLessOrEqualThan checks if ArrayLengthLessOrEqualThan validates array length is less or equal to maxValue.
func TestArrayLengthLessOrEqualThan(t *testing.T) {
	tests := []struct {
		id       int
		value    [3]int
		maxValue int
		error    bool
	}{
		{1, [3]int{1, 2, 3}, 3, false},
		{2, [3]int{1, 2, 3}, 1, true},
		{4, [3]int{1, 2, 3}, 5, false},
	}

	for _, test := range tests {
		err := ArrayLengthLessOrEqualThan(test.value, test.maxValue)

		// Check if the error result matches the expected outcome.
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}
