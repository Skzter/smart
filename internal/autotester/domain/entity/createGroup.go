package entity

// CreateGroupRequest represents the payload for creating a new group
type CreateGroupRequest struct {
	UserId      string `required,json:"userId"`
	GroupName   string `required,json:"groupName"`
	Description string `json:"description"`
}

// CreateGroupResponse represents the response after successfully creating a group
type CreateGroupResponse struct {
	GroupId string `json:"groupId"`
}
