package entity

// LocalSaveRequest represents a request to save a test locally.
type LocalSaveRequest struct {
	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
	Code   string `json:"code"`
}
