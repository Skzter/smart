package entity

import "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// Message wraps a shared Message with a Type for use in Chat entity
type Message struct {
	entity.Message
	Type Type `json:"type"`
}
