package entity

// CreateGroupRequest represents the payload for creating a new group
type CreateGroupRequest struct {
	UserID      string `json:"userId" binding:"required"`
	GroupName   string `json:"groupName" binding:"required"`
	Description string `json:"description"`
}
