package main

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

func main() {
	cfg, err := config.LoadFromPath(context.Background(), "configs/autotester.pkl")

	application.SetupRoutes(cfg)
	if err != nil {
		panic(err)
	}
}
