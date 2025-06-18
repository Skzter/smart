package assert

import (
	"testing"
)

func TestByteLength(t *testing.T) {
	tests := []struct {
		id        int
		value     []byte
		minLength int
		maxLength int
		error     bool
	}{
		{1, []byte("hello"), -1, -1, false},
		{2, []byte(""), 1, 20, true},
		{3, []byte("hello"), 1, 20, false},
		{4, []byte("hello"), 10, 20, true},
		{5, []byte("hello"), 1, 3, true},
		{7, []byte("hello"), 10, 5, true},
		{8, []byte("hello"), 5, 5, false},
		{9, []byte("hello"), 5, 10, false},
		{10, []byte("hello"), 1, 5, false},
	}

	for _, test := range tests {
		err := ByteLength(test.value, test.minLength, test.maxLength)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: '%s', minLenght: %d, maxLength: %d)", test.id, test.value, test.minLength, test.maxLength)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: '%s', minLenght: %d, maxLength: %d)", err, test.id, test.value, test.minLength, test.maxLength)
		}
	}
}

func TestStringLength(t *testing.T) {
	tests := []struct {
		id            int
		value         string
		minLength     int
		maxLength     int
		allowedValues *[]string
		error         bool
	}{
		{1, "hello", -1, -1, nil, false},
		{2, "", 1, 20, nil, true},
		{3, "hello", 1, 20, nil, false},
		{4, "hello", 10, 20, nil, true},
		{5, "hello", 1, 3, nil, true},
		{6, "hello", 1, 10, &[]string{"foo", "bar"}, true},
		{6, "hello", 1, 10, &[]string{"foo", "bar", "hello"}, false},
		{7, "hello", 10, 5, nil, true},
		{8, "hello", 5, 5, nil, false},
		{9, "hello", 5, 10, nil, false},
		{10, "hello", 1, 5, nil, false},
	}

	for _, test := range tests {
		err := StringLength(test.value, test.minLength, test.maxLength, test.allowedValues)
		if test.error == true && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: '%s', minLenght: %d, maxLength: %d, allowedValues: %s)", test.id, test.value, test.minLength, test.maxLength, *test.allowedValues)
		} else if test.error == false && err != nil {
			t.Errorf("Got error '%s', expected nil (test id: %d, value: '%s', minLenght: %d, maxLength: %d, allowedValues: %s)", err, test.id, test.value, test.minLength, test.maxLength, *test.allowedValues)
		}
	}
}

func TestStringNotEmpty(t *testing.T) {
	tests := []struct {
		id    int
		value string
	}{
		{1, ""},
		{2, "foo"},
		{3, "barfoo  "},
		{4, "a"},
	}

	for _, test := range tests {
		err := StringNotEmpty(test.value)
		if test.value == "" && err == nil {
			t.Errorf("Got nil, expected error (test id: %d, value: '%s')", test.id, test.value)
		} else if test.value != "" && err != nil {
			t.Errorf("Got error '%s', expected error (test id: %d, value: '%s')", err, test.id, test.value)
		}
	}
}

func BenchmarkStringLength(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = StringLength("hello", 1, 10, &[]string{"foo", "bar", "hello"})
	}
}
