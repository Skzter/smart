package entity

import repoEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// RequestForLLM represents a request sent to the language model, including prompts and context.
type RequestForLLM struct {
	SessionId          string `json:"conversationId"`
	RequestId          string
	SystemPrompt       *SystemPrompt
	UserPrompt         *UserPrompt
	SessionContextData *SessionSummary // data to generate context for the LLM
}

func (r RequestForLLM) ToDTO() repoEntity.Request {
	return repoEntity.Request{}
}
