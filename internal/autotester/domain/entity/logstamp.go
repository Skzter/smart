package entity

import (
	"time"

	"github.com/google/uuid"
)

type LogStamp struct {
	loggingId string
	timeStamp time.Time
	actorId   string // actorID identifies the user OR system responsible for the corresponding action
}

func NewLogStamp(actorId string) LogStamp {
	return LogStamp{
		loggingId: uuid.New().String(),
		timeStamp: time.Now(),
		actorId:   actorId,
	}
}

func (ls LogStamp) GetLoggingId() string {
	return ls.loggingId
}

func (ls LogStamp) GetTimeStamp() time.Time {
	return ls.timeStamp
}

func (ls LogStamp) GetActorId() string {
	return ls.actorId
}
