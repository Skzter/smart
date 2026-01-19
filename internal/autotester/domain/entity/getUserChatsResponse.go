package entity

// ChatSummarys is an entity for holding an array of ChatSummarys
type ChatSummarys struct {
	ChatSummarys []*ChatSummary `json:"chatSummarys"`
	HasMore      bool           `json:"hasMore"`
}
