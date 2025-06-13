package entity

type Agent string

const (
	User   Agent = "user"
	System Agent = "system"
)

type Message struct {
	Data  string `json:"data"`
	Agent Agent  `json:"agent"`
}

type ChatRequest struct {
	Message        Message `json:"message"`
	UserID         string  `json:"userId"`
	ConversationID string  `json:"conversationId"`
}

type ChatResponse struct {
	Content    ChatRequest `json:"content"`
	tools_call ToolCall
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
