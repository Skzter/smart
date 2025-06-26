package entity

// LLMResponse represents a response from the language model.
// It contains the session ID, log stamp, answer text, and test code.
type LLMResponse struct {
	SessionId
	LogStamp   LogStamp
	AnswerText *ModelAnswerText
	TestCode   *TestCode
}
