package main

import (
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

func main() {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		panic(err)
	}

	application.SetupRoutes(cfg)
}
