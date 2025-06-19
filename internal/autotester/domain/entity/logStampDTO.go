package entity

type LogStampDTO struct {
	ActorId string `json:"userId"`
}

func (LogStampDTO) ToEntity() LogStamp {
	return LogStamp{}
}
