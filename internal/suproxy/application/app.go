package application

import (
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// Run initializes the application, sets up the logger, handler, and router,
func Run() {
	logger := slog.Default()
	if logger == nil {
		panic("logger is nil")
	}

	cfg, err := config.LoadAppConfig()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	handler, err := handler.NewSuproxyController(logger, cfg)

	if err != nil {
		logger.Error("failed to create handler", "error", err)
		return
	}

	router := handler.SetupRouter()

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("failed to run server", "error", err)
		return
	}
}
