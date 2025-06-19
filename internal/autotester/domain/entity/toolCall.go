package entity

type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
}

func (ToolCall) ToDTO() ToolCallDTO {
	return ToolCallDTO{}
}
