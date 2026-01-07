package entity

// GetRemoteTestcaseRequest represents query parameters for filtering remote test cases.
// All fields are optional and validated through Gin's binding tags (UUID format and ISO8601 datetime).
type GetRemoteTestcaseRequest struct {
	Author        string `form:"author" binding:"omitempty"`
	TestcaseId    string `form:"testcaseId" binding:"omitempty,uuid"`
	CreatedAfter  string `form:"createdAfter" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CreatedBefore string `form:"createdBefore" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit         *int   `form:"limit" binding:"omitempty,min=1"`
	Offset        int    `form:"offset" binding:"omitempty,min=0"`
}
