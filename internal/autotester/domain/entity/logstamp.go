package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// LogStamp represents a log entry with a unique logging ID, timestamp, and actor ID.
// The actor ID identifies the user or system responsible for the corresponding action.
type LogStamp struct {
	loggingId string
	timeStamp time.Time
	actorId   string // actorID identifies the user OR system responsible for the corresponding action
}

// NewLogStamp creates a new LogStamp with the given actorId.
// Returns an error if the actorId is empty.
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

// GetLoggingId returns the unique logging ID of the LogStamp.
func (ls LogStamp) GetLoggingId() string {
	return ls.loggingId
}

// GetTimeStamp returns the timestamp of the LogStamp.
func (ls LogStamp) GetTimeStamp() time.Time {
	return ls.timeStamp
}

// GetActorId returns the actor ID of the LogStamp.
func (ls LogStamp) GetActorId() string {
	return ls.actorId
}

// ToDTO converts the LogStamp to a LogStampDTO.
func (LogStamp) ToDTO() LogStampDTO {
	return LogStampDTO{}
}
