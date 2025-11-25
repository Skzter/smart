package entity

import (
	"testing"
)

func TestTestStatus_String(t *testing.T) {
	tests := []struct {
		status TestStatus
		want   string
	}{
		{TestStatusPending, "pending"},
		{TestStatusNotRun, "not_run"},
		{TestStatusRunning, "running"},
		{TestStatusPassed, "passed"},
		{TestStatusFailed, "failed"},
		{TestStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("String() = %v, want %v", got, tt.want)
		}
	}
}

func TestParseTestStatus(t *testing.T) {
	for str, status := range _TestStatusValue {
		got, err := ParseTestStatus(str)
		if err != nil {
			t.Errorf("ParseTestStatus(%q) returned error: %v", str, err)
		}
		if got != status {
			t.Errorf("ParseTestStatus(%q) = %v, want %v", str, got, status)
		}
	}

	_, err := ParseTestStatus("invalid")
	if err == nil {
		t.Errorf("ParseTestStatus(\"invalid\") expected error, got nil")
	}
}

func TestIsValid(t *testing.T) {
	valid := []TestStatus{
		TestStatusPending,
		TestStatusNotRun,
		TestStatusRunning,
		TestStatusPassed,
		TestStatusFailed,
		TestStatusSkipped,
	}

	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("IsValid() = false, want true for %v", s)
		}
	}

	if TestStatus(999).IsValid() {
		t.Errorf("IsValid() = true, want false for invalid value")
	}
}

func TestScanAndValue(t *testing.T) {
	for _, status := range []TestStatus{
		TestStatusPending,
		TestStatusRunning,
		TestStatusFailed,
	} {
		val, err := status.Value()
		if err != nil {
			t.Fatalf("Value() error: %v", err)
		}

		var s TestStatus
		if err := s.Scan(val); err != nil {
			t.Fatalf("Scan() error: %v", err)
		}

		if s != status {
			t.Errorf("Scan(Value()) = %v, want %v", s, status)
		}
	}
}

func TestMarshalUnmarshalText(t *testing.T) {
	for _, status := range []TestStatus{
		TestStatusPending,
		TestStatusRunning,
		TestStatusFailed,
	} {
		data, err := status.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() error: %v", err)
		}

		var s TestStatus
		if err := s.UnmarshalText(data); err != nil {
			t.Fatalf("UnmarshalText() error: %v", err)
		}

		if s != status {
			t.Errorf("UnmarshalText(MarshalText()) = %v, want %v", s, status)
		}
	}
}
