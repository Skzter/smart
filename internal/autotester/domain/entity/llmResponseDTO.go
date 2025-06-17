package entity

type LLMResponseDTO struct {
	Text      string
	SessionID string // Response ID for conversation tracking
}

func (LLMResponseDTO) ToEntity() LLMResponse {
	return LLMResponse{}
}
