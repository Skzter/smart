//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/openai/openai-go"

	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		LoggerProvider,
		application.NewRouter,
		handler.NewSuproxyController,
	)

	return nil, nil
}

// LoggerProvider provides a new logger.
func LoggerProvider(cfg *config.Config) *slog.Logger {
	return lib.NewLogger(cfg.LogLevel)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiRepositoryProvider(logger *slog.Logger, client openai.Client, cfg *config.Config) (sharedRepo.OpenAI, error) {
	return sharedRepo.NewOpenAiRepository(logger, client, cfg.Timeout)
}
