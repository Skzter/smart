package entity

// TestStatus defines the status of a test.
// ENUM(pending, not_run, running, passed, failed, skipped)
type TestStatus int

//go:generate go-enum -f=$GOFILE --marshal
