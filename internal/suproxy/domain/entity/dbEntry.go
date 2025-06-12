package entity

type DatabaseEntry struct {
	EntryID   int  `json:"entry_id"`
	PromptID  int  `json:"prompt_id"`
	RequestID int  `json:"request_id"`
	Validated bool `json:"validated"`
}
