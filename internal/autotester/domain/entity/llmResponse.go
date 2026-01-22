package entity

// LLMResponse represents a response from the language model.
// It contains the chat ID, log stamp, answer text, and test code.
type LLMResponse struct {
	ChatId     string `json:"chatId"`
	AnswerText *ModelAnswerText
	TestCode   *TestCode
}
