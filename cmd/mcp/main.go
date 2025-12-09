package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/application/tools"
)

func main() {
	ctx := context.Background()
	cfg := &config.Config{
		Template: "Dies ist das Testgenerierungs-Template",
	}
	getTemplateTool := tools.NewGetTemplateTool(cfg)
	req := &mcp.CallToolRequest{}
	input := tools.GetTemplateInput{} // leeres Struct, bei Bedarf Felder füllen

	callRes, out, err := getTemplateTool.GetTemplate(ctx, req, input)
	if err != nil {
		fmt.Println("GetTemplate error:", err)
	} else {
		fmt.Println("CallResult:", callRes)
		fmt.Println("Template:", out.Template)
	}

	/*saveLocalService, err := service.NewTestcaseLocalStorageService(*slog logger, )
	    dockerService := handler.DockerService

		runTests := tools.NewRunTestTool(service.TestcaseLocalStorageService, tools.RunTestTool.DockerService)*/

	server := mcp.NewServer(&mcp.Implementation{Name: "check24-smart-mcp", Version: "1.0.0"}, nil)

	// register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTemplate",
		Description: "asks for generated test template"},
		func(c context.Context, req *mcp.CallToolRequest, in tools.GetTemplateInput) (*mcp.CallToolResult, tools.GetTemplateOutput, error) {
			return getTemplateTool.GetTemplate(c, req, in)
		})
	/*mcp.AddTool(server, &mcp.Tool{
		Name: "generateTest"
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "runTest"
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "getTestResult"
	})*/
}
