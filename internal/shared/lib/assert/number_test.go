package assert

import (
	"fmt"
	"testing"
)

func TestNumberBetween(t *testing.T) {
	tests := []struct {
		id       int
		value    int
		minValue int
		maxValue int
		error    bool
	}{
		{1, 10, 10, 20, false},
		{2, 15, 10, 20, false},
		{3, 20, 10, 20, false},
		{4, 5, 10, 20, true},
		{5, 25, 10, 20, true},
	}

	for _, test := range tests {
		err := NumberBetween(test.value, test.minValue, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d, maxValue: %d)", test.id, test.value, test.minValue, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d, maxValue: %d)", err, test.id, test.value, test.minValue, test.maxValue)
		}
	}
}

func TestNumberGreaterThan(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		minValue interface{}
		error    bool
	}{
		{1, 10, 10, true},
		{2, 15, 10, false},
		{3, 5, 10, true},
		{4, uint(10), uint(9), false},
		{5, float32(10), float32(9), false},
	}

	for _, test := range tests {
		err := NumberGreaterThan(test.value, test.minValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

func TestNumberGreaterOrEqualThan(t *testing.T) {
	tests := []struct {
		id       int
		value    int
		minValue int
		error    bool
	}{
		{1, 10, 10, false},
		{2, 15, 10, false},
		{4, 5, 10, true},
	}

	for _, test := range tests {
		err := NumberGreaterOrEqualThan(test.value, test.minValue)

		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, minValue: %d)", test.id, test.value, test.minValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, minValue: %d)", err, test.id, test.value, test.minValue)
		}
	}
}

func TestNumberLessThan(t *testing.T) {
	tests := []struct {
		id       int
		value    interface{}
		maxValue interface{}
		error    bool
	}{
		{1, 10, 10, true},
		{2, 15, 10, true},
		{3, 5, 10, false},
		{4, uint(9), uint(10), false},
		{5, float32(9), float32(10), false},
	}

	for _, test := range tests {
		err := NumberLessThan(test.value, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}

func TestNumberLessOrEqualThan(t *testing.T) {
	tests := []struct {
		id       int
		value    int
		maxValue int
		error    bool
	}{
		{1, 10, 10, false},
		{2, 15, 10, true},
		{4, 5, 10, false},
	}

	for _, test := range tests {
		err := NumberLessOrEqualThan(test.value, test.maxValue)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d, maxValue: %d)", test.id, test.value, test.maxValue)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d, maxValue: %d)", err, test.id, test.value, test.maxValue)
		}
	}
}

func TestId(t *testing.T) {
	tests := []struct {
		id    int
		value int
		error bool
	}{
		{1, 0, true},
		{2, 1, false},
	}

	for _, test := range tests {
		err := Id(test.value)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: %d)", test.id, test.value)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: %d)", err, test.id, test.value)
		}
	}
}

func BenchmarkNumberBetweenInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberBetween(5, 0, 10)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func BenchmarkNumberBetweenUInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberBetween(uint8(5), 0, 10)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func BenchmarkNumberBetweenFloat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberBetween(float32(5.0), 0.0, 10.0)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func BenchmarkNumberLessThanInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberLessThan(5, 10)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func BenchmarkNumberLessThanUInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberLessThan(uint8(5), 10)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func BenchmarkNumberLessThanFloat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := NumberLessThan(float32(5.0), 10.0)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func BenchmarkNumberBetweenWorst(b *testing.B) {
	var value uint32 = 5
	var min int32 = 0
	var max int32 = 10
	for n := 0; n < b.N; n++ {
		err := NumberBetween(value, min, max)
		if err != nil {
			fmt.Println(err)
		}
	}
}
