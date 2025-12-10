package entity

// GetTemplateOutput represents the result of a template retrieval request.
// Template contains the test template string that can be used as a starting point for test generation.
type GetTemplateOutput struct {
	Template string `json:"template"`
}
