package entity

type LLMResponse struct {
	SessionId
	LogStamp   LogStamp
	AnswerText *ModelAnswerText
	TestCode   *TestCode
}
