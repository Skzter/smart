package entity

type LLMResponse struct {
	SessionId  string `json:"conversationId"`
	LogStamp   LogStamp
	AnswerText *ModelAnswerText
	TestCode   *TestCode
}
