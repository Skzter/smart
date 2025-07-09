package entity

import "testing"

// TestNewLogStamp tests the NewLogStamp function.
// It checks that a LogStamp is created correctly for valid actor IDs and returns an error for empty actor IDs.
// It also verifies that each LogStamp instance has a unique logging ID.
func TestNewLogStamp(t *testing.T) {
	tests := []struct {
		TestName    string
		ActorId     string
		expectError bool
	}{
		{"simple actorId", "user123", false},
		{"another simple actorId", "system456", false},
		{"empty actorId", "", true},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			ls, err := NewLogStamp(test.ActorId)
			ls2, _ := NewLogStamp(test.ActorId)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for actorId %q, but got none", test.ActorId)
				}
				return
			}

			if err != nil {
				t.Errorf("Didn't expect error for actorId %q, but got: %v", test.ActorId, err)
			}

			if ls.GetLoggingId() == ls2.GetLoggingId() {
				t.Errorf("Expected different UUIDs, but got identical: %s", ls.GetLoggingId())
			}
		})
	}
}
