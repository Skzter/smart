package entity

// ToolCall represents a call to a tool with a name and arguments.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
