package entity

// LogStampDTO is used to transfer data stored in log stamps,
// ensuring the original LogStamp remains immutable and protected
// from external modifications.
type LogStampDTO struct {
	ActorId string `json:"userId"`
}

func (LogStampDTO) ToEntity() LogStamp {
	return LogStamp{}
}
