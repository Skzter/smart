package entity

import sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// ResponseForUser represents a response sent to the user, including text and test case(s).
type ResponseForUser struct {
	Message   sharedEntity.Message `json:"message"`
	UserId    string               `json:"userId"`
	SessionId string               `json:"conversationId"`
	ToolCall  ToolCall             `json:"tool_calls"`
	LogStamp  LogStamp             `json:"-"`
	TestCases []*TestCase          `json:"-"` // list of test cases for multiple options

}
