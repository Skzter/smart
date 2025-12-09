package tools

type GenerateTestInput struct {
	Template string `json:"template"`
}

type GenerateTestOutput struct {
	TestID     string `json:"test_id"`
	TestCode   string `json:"test_code"`
	TestStatus string `json:"test_status"`
}
