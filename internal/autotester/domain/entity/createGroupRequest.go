package entity

// CreateGroupRequest represents the payload for creating a new group
type CreateGroupRequest struct {
	UserId      string `json:"userId" binding:"required"`
	GroupName   string `json:"groupName" binding:"required"`
	Description string `json:"description"`
}
