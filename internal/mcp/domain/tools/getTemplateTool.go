package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GetTemplateTool is an amazing tool
type GetTemplateTool struct {
	logger               *slog.Logger
	autotesterAPIService service.AutotesterAPIService
}

// Creates a new TemplateTool
func NewGetTemplateTool(
	logger *slog.Logger,
	autotesterAPIService service.AutotesterAPIService,
) (*GetTemplateTool, error) {
	return &GetTemplateTool{
		logger:               logger,
		autotesterAPIService: autotesterAPIService,
	}, nil
}

// GetTemplate Requests template from backend
func (tt *GetTemplateTool) GetTemplate(ctx context.Context, request *mcp.CallToolRequest, input entity.TemplateIn) (result *mcp.CallToolResult, output entity.TemplateResponse, _ error) {
	if err := assert.NotNil(tt.logger, tt.autotesterAPIService); err != nil {
		return nil, entity.TemplateResponse{}, err
	}
	tt.logger.Log(ctx, slog.LevelInfo, "Getting template")
	return nil, tt.autotesterAPIService.GetTemplate(), nil
}
