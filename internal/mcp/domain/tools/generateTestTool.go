package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GenerateTestTool as a struct
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

// GenerateTest generates Test
func (tt *GenerateTestTool) GenerateTest(ctx context.Context,
	request *mcp.CallToolRequest,
	input entity.GenerateTestRequest) (result *mcp.CallToolResult, output entity.GenerateTestResponse, _ error) {
	tt.logger.Info("Generate Test")

	test, err := tt.autotesterAPIService.GenerateTest(ctx, &input)
	if err != nil {
		tt.logger.Error("Generate test failed")
		return nil, entity.GenerateTestResponse{}, err
	}

	tt.logger.Info("Test successfully generated", "length", len(test.Result.Body))
	return nil, *test, nil
}
