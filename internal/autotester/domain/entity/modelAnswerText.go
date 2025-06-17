package entity

type ModelAnswerText struct {
	Text string
}

func (ModelAnswerText) ToDTO() ModelAnswerTextDTO {
	return ModelAnswerTextDTO{}
}
