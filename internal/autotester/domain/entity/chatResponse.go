package entity

// ChatResponse represents a chat response containing the user request content and tool call information.
type ChatResponse struct {
	Content    UserRequest `json:"content"`
	Tools_call ToolCall    `json:"tool_call"`
}
