package entity

type ChatResponseDTO struct {
	Content    UserRequestDTO `json:"content"`
	Tools_call ToolCall       `json:"tool_call"`
}

func (ChatResponseDTO) ToEntity() ChatResponse {
	return ChatResponse{}
}
