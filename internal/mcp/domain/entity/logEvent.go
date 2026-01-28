package entity

// LogEvent represents an SSE event from the autotester backend (live test logs).
type LogEvent struct {
	Event string
	Data  string
}
