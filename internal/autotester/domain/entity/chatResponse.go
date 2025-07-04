package entity

type ChatResponse struct {
	Content    UserRequest
	Tools_call ToolCall
}

func (ChatResponse) ToDTO() ChatResponseDTO {
	return ChatResponseDTO{}
}
