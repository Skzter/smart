package entity

import sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// UserRequest represents a user request within a session, including prompt and log information.
type UserRequest struct {
	SessionId  string   `json:"conversationId"`
	LogStamp   LogStamp `json:"-"`
	UserPrompt *UserPrompt
	Message    sharedEntity.Message `json:"message"`
	UserId     string               `json:"userId"`
}
