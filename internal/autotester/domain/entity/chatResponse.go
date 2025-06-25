package entity

type ChatResponse struct {
	Content    UserRequest `json:"content"`
	Tools_call ToolCall    `json:"tool_call"`
}
