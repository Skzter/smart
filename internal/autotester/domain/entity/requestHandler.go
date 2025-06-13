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
	Tools_call ToolCall    `json:"tool_call"`
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ErrorMessage struct {
	Error string `json:"message"`
}

type Session struct {
	ConversationID string    `json:"ConversationId"`
	Messages       []Message `json:"messages"`
}

type UserResponse struct {
	UserID   string    `json:"userId"`
	Sessions []Session `json:"allConversations"`
}

type UserRequest struct {
	UserID string `json:"userId"`
}
