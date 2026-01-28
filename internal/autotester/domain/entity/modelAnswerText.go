package entity

// ModelAnswerText represents the answer jsonObject returned by the language model.
type ModelAnswerText struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}
