package main

import (
	"os"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	router, err := InitializeApp(cfg)
	if err != nil {
		logger.Error("Failed to load app config", "error", err)
		os.Exit(1)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
