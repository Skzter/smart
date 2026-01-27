package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// RunTestTool is a tool to run tests
type RunTestTool struct {
	logger               *slog.Logger
	autotesterAPIService service.AutotesterAPIService
}

// NewRunTestTool creates a new RunTestTool
func NewRunTestTool(logger *slog.Logger, autotesterAPIService service.AutotesterAPIService) (*RunTestTool, error) {
	if err := assert.NotNil(logger, autotesterAPIService); err != nil {
		return nil, err
	}
	return &RunTestTool{
		logger:               logger,
		autotesterAPIService: autotesterAPIService,
	}, nil
}

// RunTest starts to run a test in the backend
func (tt *RunTestTool) RunTest(ctx context.Context, request *mcp.CallToolRequest, input entity.ExecuteTestRequest) (result *mcp.CallToolResult, output entity.RunTestResponse, _ error) {
	tt.logger.Debug("RunTest tool called")

	testResult, err := tt.autotesterAPIService.ExecuteTest(ctx, &input)
	if err != nil {
		tt.logger.Error("Failed to run test", "error", err)
		return nil, entity.RunTestResponse{}, err
	}

	tt.logger.Debug("Test executed successfully", "result", testResult)
	return nil, entity.RunTestResponse{Result: testResult.Result}, nil
}
