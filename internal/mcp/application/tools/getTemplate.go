package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

type GetTemplateInput struct{}

type GetTemplateOutput struct {
	Template string `json:"template"`
}

type GetTemplateTool struct {
	cfg *config.Config
}

func NewGetTemplateTool(cfg *config.Config) *GetTemplateTool {
	return &GetTemplateTool{cfg: cfg}
}

func (t *GetTemplateTool) GetTemplate(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetTemplateInput,
) (*mcp.CallToolResult, GetTemplateOutput, error) {

	return nil, GetTemplateOutput{
		Template: t.cfg.Template,
	}, nil
}
