package entity

// GenerateTestInput represents the input payload for generating a test.
// Template contains the test template or instructions used to create the test code.
type GenerateTestInput struct {
	Template string `json:"template"`
}
