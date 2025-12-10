package entity

import (
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

// GetTemplateTool provides functionality to retrieve test templates from the application configuration.
// cfg holds the configuration containing the test template.
type GetTemplateTool struct {
	Cfg *config.Config
}
