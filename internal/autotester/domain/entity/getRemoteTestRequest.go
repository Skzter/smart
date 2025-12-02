package entity

// GetRemoteTestcaseRequest TODO: add godoc
type GetRemoteTestcaseRequest struct {
	Author        string `form:"author"`
	TestcaseId    string `form:"testcaseId"`
	CreatedAfter  string `form:"createdAfter"`
	CreatedBefore string `form:"createdBefore"`
}
