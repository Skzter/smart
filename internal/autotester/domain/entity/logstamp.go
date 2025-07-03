package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type LogStamp struct {
	loggingId string
	timeStamp time.Time
	ActorId   string `json:"userId"`
}

func NewLogStamp(actorId string) (LogStamp, error) {
	if actorId == "" {
		return LogStamp{}, errors.New("no actorId given")
	}
	return LogStamp{
		loggingId: uuid.New().String(),
		timeStamp: time.Now(),
		ActorId:   actorId,
	}, nil
}

func (ls LogStamp) GetLoggingId() string {
	return ls.loggingId
}

func (ls LogStamp) GetTimeStamp() time.Time {
	return ls.timeStamp
}
