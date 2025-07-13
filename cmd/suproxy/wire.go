//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/openai/openai-go"
	oaoption "github.com/openai/openai-go/option"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/logger"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		LoggerProvider,
		application.NewRouter,
		handler.NewSuproxyController,
		service.NewValidator,
		HTTPClientProvider,
		OpenAiServiceProvider,
		OpenAiRepositoryProvider,
		OpenAIClientProvider,
	)

	return nil, nil
}

// LoggerProvider provides a new logger.
func LoggerProvider(cfg *config.Config) *slog.Logger {
	return logger.NewLogger(cfg.LogLevel)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiRepositoryProvider(logger *slog.Logger, client openai.Client, cfg *config.Config) (sharedRepo.OpenAI, error) {
	return sharedRepo.NewOpenAiRepository(logger, client, cfg.Timeout)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiServiceProvider(logger *slog.Logger, repo sharedRepo.OpenAI) (sharedService.OpenAI, error) {
	return sharedService.NewOpenAIService(logger, repo)
}

func HTTPClientProvider() *http.Client {
	return &http.Client{}
}

func OpenAIClientProvider() openai.Client {
	return openai.NewClient(oaoption.WithAPIKey(build.OpenAIKey))
}
