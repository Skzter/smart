package entity

// ResponseForUser represents a response sent to the user, including text and test case(s).
type ResponseForUser struct {
	Message   Message     `json:"message"`
	UserId    string      `json:"userId"`
	SessionId string      `json:"conversationId"`
	ToolCall  ToolCall    `json:"tool_calls"`
	LogStamp  LogStamp    `json:"-"`
	TestCases []*TestCase `json:"-"` // list of test cases for multiple options

}
