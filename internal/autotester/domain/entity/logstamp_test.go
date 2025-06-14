package entity

import "testing"

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

			if ls.GetActorId() != test.ActorId {
				t.Errorf("Expected ActorID %s, got %s", test.ActorId, ls.GetActorId())
			}
			if ls2.GetActorId() != test.ActorId {
				t.Errorf("Expected ActorID %s, got %s", test.ActorId, ls.GetActorId())
			}

			if ls.GetLoggingId() == ls2.GetLoggingId() {
				t.Errorf("Expected different UUIDs, but got identical: %s", ls.GetLoggingId())
			}
		})
	}
}
