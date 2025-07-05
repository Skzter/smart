package assert

import (
	"testing"
)

// TestMapLengthBetween tests MapLengthBetween.
// It checks that MapLengthBetween returns an error if the map length is outside the given range.
func TestMapLengthBetween(t *testing.T) {
	data := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	tests := []struct {
		id       int
		value    map[int]bool
		minValue int
		maxValue int
		error    bool
	}{
		{1, data, 1, 5, false},
		{2, data, 1, 3, false},
		{3, data, 3, 5, false},
		{4, data, 1, 2, true},
		{5, data, 4, 5, true},
		{6, data, 2, 1, true},
	}

	for _, test := range tests {
		err := MapLengthBetween(test.value, test.minValue, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %v, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %v, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}
}

// TestMapLengthGreaterThan tests MapLengthGreaterThan.
// It checks that MapLengthGreaterThan returns an error if the map length is not greater than minValue.
func TestMapLengthGreaterThan(t *testing.T) {
	data := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	tests := []struct {
		id       int
		value    map[int]bool
		minValue int
		error    bool
	}{
		{1, data, 3, true},
		{2, data, 1, false},
		{4, data, 5, true},
	}

	for _, test := range tests {
		err := MapLengthGreaterThan(test.value, test.minValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %v, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %v, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

// TestMapLengthGreaterOrEqualThan tests MapLengthGreaterOrEqualThan.
// It checks that MapLengthGreaterOrEqualThan returns an error if the map length is not greater or equal to minValue.
func TestMapLengthGreaterOrEqualThan(t *testing.T) {
	data := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	tests := []struct {
		id       int
		value    map[int]bool
		minValue int
		error    bool
	}{
		{1, data, 3, false},
		{2, data, 1, false},
		{4, data, 5, true},
	}

	for _, test := range tests {
		err := MapLengthGreaterOrEqualThan(test.value, test.minValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %v, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %v, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

// TestMapLengthLessThan tests MapLengthLessThan.
// It checks that MapLengthLessThan returns an error if the map length is not less than maxValue.
func TestMapLengthLessThan(t *testing.T) {
	data := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	tests := []struct {
		id       int
		value    map[int]bool
		maxValue int
		error    bool
	}{
		{1, data, 3, true},
		{2, data, 1, true},
		{4, data, 5, false},
	}

	for _, test := range tests {
		err := MapLengthLessThan(test.value, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %v, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %v, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}

// TestMapLengthLessOrEqualThan tests MapLengthLessOrEqualThan.
// It checks that MapLengthLessOrEqualThan returns an error if the map length is not less or equal to maxValue.
func TestMapLengthLessOrEqualThan(t *testing.T) {
	data := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	tests := []struct {
		id       int
		value    interface{}
		maxValue int
		error    bool
	}{
		{1, data, 3, false},
		{2, data, 1, true},
		{3, 5, 1, true},
		{4, data, 5, false},
	}

	for _, test := range tests {
		err := MapLengthLessOrEqualThan(test.value, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %v, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %v, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}
