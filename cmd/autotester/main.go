package main

import "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	router, err := InitializeApp(cfg)
	if err != nil {
		panic(err)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
