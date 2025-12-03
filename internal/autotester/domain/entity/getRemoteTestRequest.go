package entity

// GetRemoteTestcaseRequest TODO: add godoc
type GetRemoteTestcaseRequest struct {
	Author        string `form:"author" binding:"omitempty,uuid"`
	TestcaseId    string `form:"testcaseId" binding:"omitempty,uuid"`
	CreatedAfter  string `form:"createdAfter" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CreatedBefore string `form:"createdBefore" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
