package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type LogStamp struct {
	loggingId string
	timeStamp time.Time
	actorId   string // actorID identifies the user OR system responsible for the corresponding action
}

func NewLogStamp(actorId string) (LogStamp, error) {
	if actorId == "" {
		return LogStamp{}, errors.New("no actorId given")
	}
	return LogStamp{
		loggingId: uuid.New().String(),
		timeStamp: time.Now(),
		actorId:   actorId,
	}, nil

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
