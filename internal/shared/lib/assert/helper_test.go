//nolint:dupl,goconst,gocritic
package assert

import (
	"reflect"
	"testing"
)

// TestCheckInt tests the checkNumber function with various int types.
// It verifies correct error handling for different value, min, and max combinations.
func TestCheckInt(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		minValue interface{}
		maxValue interface{}
		error    bool
	}{
		{1, int8(5), 10, 1, true},
		{2, int8(0), 1, 10, true},
		{3, int8(1), 1, 10, false},
		{4, int8(5), 1, 10, false},
		{5, int8(10), 1, 10, false},
		{6, int8(15), 1, 10, true},
		{7, int8(10), 10, 10, false},
		{8, int8(10), float64(10), 10, true},
		{9, int8(10), 10, float64(10), true},
	}

	// int
	for _, test := range tests {
		err := checkNumber(int(test.value.(int8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%v', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// int8
	for _, test := range tests {
		err := checkNumber(test.value.(int8), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// int16
	for _, test := range tests {
		err := checkNumber(int16(test.value.(int8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// int32
	for _, test := range tests {
		err := checkNumber(int32(test.value.(int8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// int64
	for _, test := range tests {
		err := checkNumber(int64(test.value.(int8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// non integer
	valueString := "hello"
	var minValue interface{} = 1
	var maxValue interface{} = 10

	err := checkNumber(valueString, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%s', minValue: %d, maxValue: %d)", valueString, minValue, maxValue)
	}

	valueUint := uint(4)
	err = checkNumber(valueUint, &minValue, &maxValue, false)
	if err != nil {
		t.Errorf("Got error '%s', expected nil (value: %d, minValue: %d, maxValue: %d)", err, valueUint, minValue, maxValue)
	}

	valueFloat := 5.5
	err = checkNumber(valueFloat, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%f', minValue: %d, maxValue: %d)", valueFloat, minValue, maxValue)
	}
}

// TestCheckUInt tests the checkNumber function with various uint types.
// It verifies correct error handling for different value, min, and max combinations.
func TestCheckUInt(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		minValue interface{}
		maxValue interface{}
		error    bool
	}{
		{1, uint8(5), uint(10), uint(1), true},
		{2, uint8(0), uint(1), uint(10), true},
		{3, uint8(1), uint(1), uint(10), false},
		{4, uint8(5), uint(1), uint(10), false},
		{5, uint8(10), uint(1), uint(10), false},
		{6, uint8(15), uint(1), uint(10), true},
		{7, uint8(10), uint(10), uint(10), false},
		{8, uint8(10), float64(10), uint(10), true},
		{9, uint8(10), uint(10), float64(10), true},
	}

	// uint
	for _, test := range tests {
		err := checkNumber(uint(test.value.(uint8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// uint8
	for _, test := range tests {
		err := checkNumber(test.value.(uint8), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// uint16
	for _, test := range tests {
		err := checkNumber(uint16(test.value.(uint8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// uint32
	for _, test := range tests {
		err := checkNumber(uint32(test.value.(uint8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// uint64
	for _, test := range tests {
		err := checkNumber(uint64(test.value.(uint8)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// non integer
	valueString := "hello"
	var minValue interface{} = uint(1)
	var maxValue interface{} = uint(10)

	err := checkNumber(valueString, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%s', minValue: %d, maxValue: %d)", valueString, minValue, maxValue)
	}

	valueInt := 4
	err = checkNumber(valueInt, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%d', minValue: %d, maxValue: %d)", valueInt, minValue, maxValue)
	}

	valueFloat := 5.5
	err = checkNumber(valueFloat, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%f', minValue: %d, maxValue: %d)", valueFloat, minValue, maxValue)
	}
}

// TestCheckFloat tests the checkNumber function with float types.
// It checks error handling for different float values and boundaries.
func TestCheckFloat(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		minValue interface{}
		maxValue interface{}
		error    bool
	}{
		{1, 5., 10., 1., true},
		{2, 0., 1., 10., true},
		{3, 1., 1., 10., false},
		{4, 5., 1., 10., false},
		{5, 10., 1., 10., false},
		{6, 15., 1., 10., true},
		{7, 10., 10., 10., false},
		{8, 10., int8(10), 10., true},
		{9, 10., 10., int8(10), true},
	}

	// float32
	for _, test := range tests {
		err := checkNumber(float32(test.value.(float64)), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %f, minValue: %f, maxValue: %f)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %f, minValue: %f, maxValue: %f)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// float64
	for _, test := range tests {
		err := checkNumber(test.value.(float64), &test.minValue, &test.maxValue, false)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %f, minValue: %f, maxValue: %f)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %f, minValue: %f, maxValue: %f)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// non float
	valueString := "hello"
	var minValue interface{} = float64(1)
	var maxValue interface{} = float64(10)

	err := checkNumber(valueString, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%s', minValue: %f, maxValue: %f)", valueString, minValue, maxValue)
	}

	valueInt := 4
	err = checkNumber(valueInt, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%d', minValue: %f, maxValue: %f)", valueInt, minValue, maxValue)
	}

	valueUInt := uint(5)
	err = checkNumber(valueUInt, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%d', minValue: %f, maxValue: %f)", valueUInt, minValue, maxValue)
	}
}

// TestCheckClosedBoundaries tests checkNumber with closed boundaries enabled.
// It ensures correct error handling when values are at the boundary.
func TestCheckClosedBoundaries(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		minValue interface{}
		maxValue interface{}
		error    bool
	}{
		{1, 10, 0, 10, true},
		{2, 0, 0, 10, true},
		{3, uint(10), 0, 10., true},
		{4, uint(0), 0, 10., true},
		{5, 10., 0., 10., true},
		{6, 0., 0., 10., true},
	}

	for _, test := range tests {
		err := checkNumber(test.value, &test.minValue, &test.maxValue, true)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %f, minValue: %f, maxValue: %f)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %f, minValue: %f, maxValue: %f)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}

	// non float
	valueString := "hello"
	var minValue interface{} = float64(1)
	var maxValue interface{} = float64(10)

	err := checkNumber(valueString, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%s', minValue: %f, maxValue: %f)", valueString, minValue, maxValue)
	}

	valueInt := 4
	err = checkNumber(valueInt, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%d', minValue: %f, maxValue: %f)", valueInt, minValue, maxValue)
	}

	valueUInt := uint(5)
	err = checkNumber(valueUInt, &minValue, &maxValue, false)
	if err == nil {
		t.Errorf("Got nil, expected error (value: '%d', minValue: %f, maxValue: %f)", valueUInt, minValue, maxValue)
	}
}

// BenchmarkCheckNumberInt benchmarks checkNumber with int values.
func BenchmarkCheckNumberInt(b *testing.B) {
	var minValue interface{} = 1
	var maxValue interface{} = 10
	var value = 10

	for n := 0; n < b.N; n++ {
		_ = checkNumber(value, &minValue, &maxValue, false)
	}
}

// BenchmarkCheckNumberUInt benchmarks checkNumber with uint values.
func BenchmarkCheckNumberUInt(b *testing.B) {
	var minValue interface{} = 1
	var maxValue interface{} = 10
	var value uint = 10

	for n := 0; n < b.N; n++ {
		_ = checkNumber(value, &minValue, &maxValue, false)
	}
}

// BenchmarkCheckNumberFloat benchmarks checkNumber with float values.
func BenchmarkCheckNumberFloat(b *testing.B) {
	var minValue interface{} = 1.0
	var maxValue interface{} = 10.0
	var value float32 = 10

	for n := 0; n < b.N; n++ {
		err := checkNumber(value, &minValue, &maxValue, false)
		if err != nil {
			println(err)
		}
	}
}

// TestArrayLength tests the arrayLength function.
// It checks correct length calculation and error handling for various types.
func TestArrayLength(t *testing.T) {
	tests := []struct {
		id     int
		value  interface{}
		length int
		error  bool
	}{
		{1, 0, 0, true},
		{2, "hello", 0, true},
		{3, [2]int{1, 2}, 2, false},
		{4, [3]string{"a", "b", "c"}, 3, false},
		{5, []int{}, 0, false},
		{6, []struct{ id int }{{1}, {2}, {3}}, 3, false},
	}

	for _, test := range tests {
		length, err := arrayLength(test.value)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value type: %s, lenght: %d)", test.id, reflect.TypeOf(test.value).Kind(), length)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value type: %s, length: %d)", err, test.id, reflect.TypeOf(test.value).Kind(), length)
		} else if test.length != length {
			t.Errorf("Got wrong length %d, expected %d", length, test.length)
		}
	}
}

// BenchmarkArrayLength benchmarks the arrayLength function.
func BenchmarkArrayLength(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = arrayLength([]struct{ id int }{{1}, {2}, {3}})
	}
}

// TestCheckArrayLength tests the checkArrayLength function.
// It verifies correct error handling for different array values and boundaries.
func TestCheckArrayLength(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
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
		{7, 5, 1, 5, true},
	}

	for _, test := range tests {
		err := checkArrayLength(test.value, &test.minValue, &test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}
}

// TestMapLength tests the arrayLength function with map-like and array types.
// It checks correct length calculation and error handling.
func TestMapLength(t *testing.T) {
	tests := []struct {
		id     int
		value  interface{}
		length int
		error  bool
	}{
		{1, 0, 0, true},
		{2, "hello", 0, true},
		{3, [2]int{1, 2}, 2, false},
		{4, [3]string{"a", "b", "c"}, 3, false},
		{5, []int{}, 0, false},
		{6, []struct{ id int }{{1}, {2}, {3}}, 3, false},
	}

	for _, test := range tests {
		length, err := arrayLength(test.value)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value type: %s, lenght: %d)", test.id, reflect.TypeOf(test.value).Kind(), length)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value type: %s, length: %d)", err, test.id, reflect.TypeOf(test.value).Kind(), length)
		} else if test.length != length {
			t.Errorf("Got wrong length %d, expected %d", length, test.length)
		}
	}
}
