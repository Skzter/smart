package main

import (
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/tracing"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	tracer, shutdown, err := tracing.Setup("autotester")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := shutdown(); err != nil {
			panic(err)
		}
	}()

	router, err := InitializeApp(cfg, tracer)
	if err != nil {
		panic(err)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
