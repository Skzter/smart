package entity

// Chats is an entity for holding an array of ChatSummarys
type Chats struct {
	Chats []*ChatSummary `json:"chats"`
}
