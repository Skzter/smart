package entity

// LogStampDTO is a data transfer object for LogStamp.
// It contains the actor ID as a string.
type LogStampDTO struct {
	ActorId string `json:"userId"`
}

// ToEntity converts the LogStampDTO to a LogStamp entity.
// Returns an empty LogStamp.
func (LogStampDTO) ToEntity() LogStamp {
	return LogStamp{}
}
