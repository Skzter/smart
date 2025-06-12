package entity

type LLMResponse struct {
	SessionId
	logStamp   LogStamp
	answerText *ModelAnswerText
	testCode   *TestCode
}
