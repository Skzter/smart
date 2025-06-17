package entity

type ModelAnswerTextDTO struct {
	Text string `json:"data"`
}

func (ModelAnswerTextDTO) ToEntity() ModelAnswerText {
	return ModelAnswerText{}
}
