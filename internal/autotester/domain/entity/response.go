package entity

type LLMResponse struct {
	sessionID  SessionID
	logStamp   LogStamp
	answerText ModelAnswerText
	testCode   TestCode
}

type ResponseForUser struct {
	sessionID    SessionID
	logStamp     LogStamp
	responseText ModelAnswerText
	testCases    []TestCase
	uiViewUpdate UiViewUpdate
}

type ModelAnswerText struct {
	text string
}

type UiViewUpdate struct {
	url         string
	elementName string
}
