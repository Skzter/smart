package entity

import "time"

// Group represents a collection of chats with metadata about creation and ownership.
// It contains identifying information, descriptive details, and tracks when and by whom it was created.
type Group struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"descriprion"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
	Chats       []string  `json:"chats"`
}
