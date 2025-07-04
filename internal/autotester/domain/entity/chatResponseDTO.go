package entity

// ChatResponseDTO is a data transfer object for chat responses.
// It contains the user request content and tool call information.
type ChatResponseDTO struct {
	Content    UserRequestDTO `json:"content"`
	Tools_call ToolCall       `json:"tool_call"`
}

// ToEntity converts the ChatResponseDTO to a ChatResponse entity.
// Returns an empty ChatResponse.
func (ChatResponseDTO) ToEntity() ChatResponse {
	return ChatResponse{}
}
