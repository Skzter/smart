package entity

// ToolCall represents a call to a tool with a name and arguments.
type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
}

// ToDTO converts the ToolCall to a ToolCallDTO.
// Returns an empty ToolCallDTO.
func (ToolCall) ToDTO() ToolCallDTO {
	return ToolCallDTO{}
}
