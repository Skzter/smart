package entity

// GenerateTestToolResponse represents the combined result of a test generation request.
// It follows a "Sealed Union" pattern because a request can have two mutually exclusive outcomes:
// 1. A validation feedback (ValidateMsg) if the prompt was incomplete or ambiguous.
// 2. A generated test (GenerateMsg) if the prompt was valid and processed.
//
// By using pointers with 'omitempty', we ensure that the JSON response only contains the
// field that was actually populated. This allows the calling LLM to easily distinguish
// between requested feedback and the final result.
type GenerateTestToolResponse struct {
	// ValidateMsg contains feedback from the validation stage.
	// If present, the prompt needs refinement.
	ValidateMsg *ValidateMessage `json:"validateMessage,omitempty"`

	// GenerateMsg contains the successfully generated test code.
	// If present, the generation was successful.
	GenerateMsg *GenerateMessage `json:"generateMessage,omitempty"`

	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
}
