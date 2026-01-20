package entity

// ExecuteTestRequest represents an instruction sent to the MCP to
// execute a previously generated test. Add fields (e.g. TestID,
// ChatID, Options) when the executor interface requires them.
type ExecuteTestRequest struct {
	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
	Test   string `json:"test"`
}
