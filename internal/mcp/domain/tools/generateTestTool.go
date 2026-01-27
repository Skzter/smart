package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GenerateTestTool is a tool that creates a new test from the provided specification after validation
type GenerateTestTool struct {
	logger               *slog.Logger
	autotesterAPIService service.AutotesterAPIService
}

// NewGenerateTestTool creates GenerateTestTool
func NewGenerateTestTool(logger *slog.Logger, autotesterAPIService service.AutotesterAPIService) (*GenerateTestTool, error) {
	if err := assert.NotNil(logger, autotesterAPIService); err != nil {
		return nil, err
	}
	return &GenerateTestTool{
		logger:               logger,
		autotesterAPIService: autotesterAPIService,
	}, nil
}

// GenerateTest generates Test with validation
func (tt *GenerateTestTool) GenerateTest(ctx context.Context,
	request *mcp.CallToolRequest,
	input entity.GenerateTestRequest) (result *mcp.CallToolResult, output entity.GenerateTestToolResponse, _ error) {
	tt.logger.Debug("Generate Test with validation")

	resp, err := tt.autotesterAPIService.GenerateTest(ctx, &input)
	if err != nil {
		tt.logger.Error("Generate test failed")
		return nil, entity.GenerateTestToolResponse{}, err
	}

	return nil, *resp, nil
}
