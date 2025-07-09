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
	ActorId   string `json:"userId"`
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
		ActorId:   actorId,
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
