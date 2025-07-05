package entity

// ToolCallDTO is a data transfer object for ToolCall.
// It contains the tool name and its arguments.
type ToolCallDTO struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToEntity converts the ToolCallDTO to a ToolCall entity.
// Returns an empty ToolCall.
func (t ToolCallDTO) ToEntity() ToolCall {
	return ToolCall{}
}
