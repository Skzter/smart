package entity

import sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// ResponseForUser represents a response sent to the user, including text and test case(s).
type ResponseForUser struct {
	Message   sharedEntity.Message `json:"message"`
	UserId    string               `json:"userId"`
	ChatId    string               `json:"chatId"`
	Title     string               `json:"title,omitempty"`
	ToolCall  ToolCall             `json:"tool_calls"`
	TestCases []*TestCase          `json:"-"` // list of test cases for multiple options
}
