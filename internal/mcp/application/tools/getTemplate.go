package tools

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

// GetTemplateInput is the input payload for retrieving a test template.
// It contains no parameters as the template is determined by the server configuration.
type GetTemplateInput struct{}

// GetTemplateOutput represents the result of a template retrieval request.
// Template contains the test template string that can be used as a starting point for test generation.
type GetTemplateOutput struct {
	Template string `json:"template"`
}

// GetTemplateTool provides functionality to retrieve test templates from the application configuration.
// cfg holds the configuration containing the test template.
type GetTemplateTool struct {
	cfg *config.Config
}

// NewGetTemplateTool creates a new GetTemplateTool instance with the provided configuration.
// It returns a pointer to the initialized GetTemplateTool.
func NewGetTemplateTool(cfg *config.Config) *GetTemplateTool {
	return &GetTemplateTool{
		cfg: cfg,
	}
}

// Execute returns the test template from the configuration.
// It ignores the input context and input parameters and always returns the configured template.
// Returns a GetTemplateOutput containing the template or an error if execution fails.
func (t *GetTemplateTool) Execute(ctx context.Context, in GetTemplateInput) (*GetTemplateOutput, error) {
	return &GetTemplateOutput{
		Template: t.cfg.TestTemplate,
	}, nil
}
