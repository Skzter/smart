package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

func TestGetTemplateTool_GetTemplate(t *testing.T) {
	// Arrange
	expectedTemplate := "Hello from template"
	cfg := &config.Config{
		Template: expectedTemplate,
	}

	tool := NewGetTemplateTool(cfg)
	ctx := context.Background()

	// Act
	result, output, err := tool.GetTemplate(ctx, &mcp.CallToolRequest{}, GetTemplateInput{})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected result to be nil, got %v", result)
	}

	if output.Template != expectedTemplate {
		t.Errorf("expected template %q, got %q", expectedTemplate, output.Template)
	}
}
