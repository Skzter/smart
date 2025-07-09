package entity

// ResponseForUser represents a response sent to the user, including text and test case(s).
type ResponseForUser struct {
	SessionId
	LogStamp     LogStamp
	ResponseText ModelAnswerText
	TestCases    []*TestCase // list of test cases for multiple options
}

// ToDTO converts the ResponseForUser to a ResponseForUserDTO.
// Returns an empty ResponseForUserDTO.
func (ResponseForUser) ToDTO() ResponseForUserDTO {
	return ResponseForUserDTO{}
}
