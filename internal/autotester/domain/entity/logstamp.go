package entity

import "time"

// LogStamp represents a timestamped log entry containing metadata such as the logging ID and the user ID.
type LogStamp struct {
	loggingID string    //loggingID uniquely identifies each logStamp
	timeStamp time.Time //timeStamp for the when
	userID    string    //userID identifies the user or system responsible for the corresponding action
}

func (ls *LogStamp) GetLoggingId() string {
	return ls.loggingID
}

func (ls *LogStamp) GetTimeStamp() time.Time {
	return ls.timeStamp
}

func (ls *LogStamp) GetUserID() string {
	return ls.userID
}
