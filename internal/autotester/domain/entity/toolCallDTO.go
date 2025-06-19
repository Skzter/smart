package entity

type ToolCallDTO struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (t ToolCallDTO) ToEntity() ToolCall {
	return ToolCall{}
}
