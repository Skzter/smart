package entity

// ChatResponse represents a chat response containing the user request content and tool call information.
type ChatResponse struct {
	Content    UserRequest
	Tools_call ToolCall
}

// ToDTO converts the ChatResponse to a ChatResponseDTO.
// Returns an empty ChatResponseDTO.
func (ChatResponse) ToDTO() ChatResponseDTO {
	return ChatResponseDTO{}
}
