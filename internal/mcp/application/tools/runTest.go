package tools

import (
	"context"
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

type RunTestTool struct {
	SaveLocalService service.TestcaseLocalStorageService
	DockerService    service.Docker
}

func NewRunTestTool(saveLocal service.TestcaseLocalStorageService, docker service.Docker) *RunTestTool {
	return &RunTestTool{
		SaveLocalService: saveLocal,
		DockerService:    docker,
	}
}

type RunTestInput struct {
	UserID    string `json:"userId"`
	TestId    string `json:"testId"`
	SessionID string `json:"sessionId"`
}

type RunTestOutput struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (t *RunTestTool) Execute(ctx context.Context, in RunTestInput) (*RunTestOutput, error) {

	// 1. Eingaben prüfen
	if in.UserID == "" || in.TestId == "" || in.SessionID == "" {
		return &RunTestOutput{
			Success: false,
			Error:   "missing required parameters",
		}, nil
	}

	// 2. Testfile finden
	testfile, err := t.SaveLocalService.GetTestPath(in.TestId, in.UserID, in.SessionID)
	if err != nil {
		return &RunTestOutput{
			Success: false,
			Error:   fmt.Sprintf("file not available: %s", err.Error()),
		}, nil
	}

	// 3. Test ausführen
	if err := t.DockerService.RunTest(nil, testfile); err != nil {
		return &RunTestOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 4. Ausgabe lesen
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
