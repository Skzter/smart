package entity

// GenerateTestResponse represents the response payload returned by the
// Autotester/chat backend when a test generation request was performed.
type GenerateTestResponse struct {
	Result GenerateMessage `json:"message"`
	UserId string          `json:"userId"`
	ChatId string          `json:"chatId"`
}
