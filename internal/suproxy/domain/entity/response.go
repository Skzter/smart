package entity

import "time"

type Response struct {
	ResponseID      int       `json:"response_id"`
	RequestID       int       `json:"request_id"`
	Timestamp       time.Time `json:"timestamp"`
	ResponseContent string    `json:"response_content"`
	Source          string    `json:"source"`
}
