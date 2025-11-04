package entity

// LlmValidationResponse is for the validation response from the llm
type LlmValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
