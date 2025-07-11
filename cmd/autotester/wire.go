//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/openai/openai-go"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		shared.SharedProviderSet,
		LoggerProvider,
		OpenAiRepositoryProvider,
		application.NewRouter,
		handler.NewAutotesterController,
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

// SessionSummaryParquetWrapperProvider provides a new session summary parquet wrapper.
func SessionSummaryParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig) (wrapperService.ParquetFileWrapper[entity.SessionSummary], error) {
	return wrapperService.NewParquetWrapper[entity.SessionSummary](logger, cfg)
}

// TestCaseParquetWrapperProvider provides a new test case parquet wrapper.
func TestCaseParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig) (wrapperService.ParquetFileWrapper[entity.TestCase], error) {
	return wrapperService.NewParquetWrapper[entity.TestCase](logger, cfg)
}
