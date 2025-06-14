package entity_test

import (
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

func TestNewLogStamp(t *testing.T) {

	actorID := "user123"
	ls := entity.NewLogStamp(actorID)
	ls2 := entity.NewLogStamp(actorID)

	if ls.GetActorId() != actorID {
		t.Errorf("Expected ActorID %s, got %s", actorID, ls.GetActorId())
	}
	if ls2.GetActorId() != actorID {
		t.Errorf("Expected ActorID %s, got %s", actorID, ls.GetActorId())
	}

	if ls.GetLoggingId() == ls2.GetLoggingId() {
		t.Errorf("Expected different UUIDs, but got identical: %s", ls.GetLoggingId())
	}

}
