package entity

type ResponseForUser struct {
	SessionId
	logStamp     LogStamp
	responseText ModelAnswerText
	testCases    []TestCase
}
