package entity

import (
	"time"
)

type LogStempDTO struct {
	LoggingID string    `gorm:"primaryKey;type:uuid;column:logging_id"` // loggingID uniquely identifies each logStamp
	TimeStamp time.Time // timeStamp for the when
	UserID    string    // userID identifies the user or system responsible for the corresponding action
}
