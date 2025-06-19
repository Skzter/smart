package entity

// ResponseForUser represents a response sent to the user, including text and test case(s).
type ResponseForUser struct {
	SessionId
	LogStamp     LogStamp
	ResponseText ModelAnswerText
	TestCases    []*TestCase // list of test cases for multiple options
}

func (ResponseForUser) ToDTO() ResponseForUserDTO {
	return ResponseForUserDTO{}
}
