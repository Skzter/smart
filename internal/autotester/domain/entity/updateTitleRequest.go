package entity

// UpdateTitleRequest represents a request to update the title of a chat.
type UpdateTitleRequest struct {
	Title string `json:"title" binding:"required"`
}
