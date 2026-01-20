package entity

// SaveTestRequest represents a request to persist a generated test
// locally or in a storage backend. Extend with fields like TestID,
// Code, UserID when persistence is implemented.
type SaveTestRequest struct {
	Code   string `json:"code"`
	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
}
