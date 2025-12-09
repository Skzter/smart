package tools

import (
	"context"
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

// RunTestTool coordinates locating and executing test files.
// SaveLocalService provides access to local test files.
// DockerService runs tests inside a Docker container and reads logs.
type RunTestTool struct {
	SaveLocalService service.TestcaseLocalStorageService
	DockerService    service.Docker
}

// NewRunTestTool creates a new RunTestTool with the given local storage and docker services.
func NewRunTestTool(saveLocal service.TestcaseLocalStorageService, docker service.Docker) *RunTestTool {
	return &RunTestTool{
		SaveLocalService: saveLocal,
		DockerService:    docker,
	}
}

// RunTestInput is the input payload for executing a test.
// UserID is the identifier of the user requesting the run.
// TestId is the identifier of the test to run.
// SessionID is the identifier of the session.
type RunTestInput struct {
	UserID    string `json:"userId"`
	TestId    string `json:"testId"`
	SessionID string `json:"sessionId"`
}

// RunTestOutput represents the result of a test run.
// Success indicates whether the run completed successfully.
// Result holds the test output when successful.
// Error holds a human-readable error message when not successful.
type RunTestOutput struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Execute validates the input, locates the test file, runs the test via Docker and returns the run output.
// On operational failures a RunTestOutput with Success=false and a descriptive Error is returned.
// The function returns an error only for unexpected internal failures.
func (t *RunTestTool) Execute(ctx context.Context, in RunTestInput) (*RunTestOutput, error) {

	if in.UserID == "" || in.TestId == "" || in.SessionID == "" {
		return &RunTestOutput{
			Success: false,
			Error:   "missing required parameters",
		}, nil
	}

	testfile, err := t.SaveLocalService.GetTestPath(in.TestId, in.UserID, in.SessionID)
	if err != nil {
		return &RunTestOutput{
			Success: false,
			Error:   fmt.Sprintf("file not available: %s", err.Error()),
		}, nil
	}

	if err := t.DockerService.RunTest(nil, testfile); err != nil {
		return &RunTestOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	output, err := t.DockerService.ReadLog(testfile)
	if err != nil {
		return &RunTestOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &RunTestOutput{
		Success: true,
		Result:  output,
	}, nil
}
