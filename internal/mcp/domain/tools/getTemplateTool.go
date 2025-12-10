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

// NewGetTemplateTool creates a new TemplateTool
func NewGetTemplateTool(logger *slog.Logger, autotesterAPIService service.AutotesterAPIService) (*GetTemplateTool, error) {
	if err := assert.NotNil(logger, autotesterAPIService); err != nil {
		return nil, err
	}
	return &GetTemplateTool{
		logger:               logger,
		autotesterAPIService: autotesterAPIService,
	}, nil
}

// GetTemplate Requests template from backend
func (tt *GetTemplateTool) GetTemplate(ctx context.Context, request *mcp.CallToolRequest, input entity.TemplateIn) (result *mcp.CallToolResult, output entity.TemplateResponse, _ error) {
	tt.logger.Info("GetTemplate tool called")

	template, err := tt.autotesterAPIService.GetTemplate(ctx)
	if err != nil {
		tt.logger.Error("Failed to get template", "error", err)
		return nil, entity.TemplateResponse{}, err
	}

	tt.logger.Info("Template retrieved successfully", "length", len(template.Content))
	return nil, entity.TemplateResponse{Content: template.Content}, nil
}
