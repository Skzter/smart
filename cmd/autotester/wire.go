//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/logger"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		shared.SharedProviderSet,
		LoggerProvider,
		OpenAiRepositoryProvider,
		FileSystemProvider,
		repository.NewTestcaseLocalStorageRepository,
		service.NewValidatePromptService,
		TestcaseLocalStorageServiceProvider,
		application.NewRouter,
		handler.NewAutotesterController,
		TagListParquetWrapperProvider,
		service.NewGeneratePromptService,
		S3WrapperProvider,
		TaglistStorageProvider,
	)

	return nil, nil
}

// LoggerProvider provides a new logger.
func LoggerProvider(cfg *config.Config) *slog.Logger {
	return logger.NewLogger(cfg.LogLevel)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiRepositoryProvider(logger *slog.Logger, client sharedRepo.OpenAIClient, cfg *config.Config) (sharedRepo.OpenAI, error) {
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

func TagListParquetWrapperProvider(logger *slog.Logger) (wrapperService.ParquetFileWrapper[sharedEntity.TagList], error) {
	return wrapperService.NewParquetWrapper[sharedEntity.TagList](logger, wrapperService.DefaultParquetConfig())
}

func S3WrapperProvider(logger *slog.Logger, cfg *config.Config) (wrapperService.S3StorageWrapper, error) {
	config := wrapperEntity.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapperService.NewS3Wrapper(logger, config)
}

func TaglistStorageProvider(
	logger *slog.Logger,
	cfg *config.Config,
	s3wrapper wrapperService.S3StorageWrapper,
	parquetWrapper wrapperService.ParquetFileWrapper[sharedEntity.TagList],
) (sharedRepo.TaglistStorage, error) {
	return sharedRepo.NewTaglistStorage(
		logger,
		s3wrapper,
		parquetWrapper,
		cfg.TaglistPrefix)
}

// FileSystemProvider provides a new filesystem.
func FileSystemProvider(cfg *config.Config) (repository.FileSystem, error) {
	return repository.NewOSFileSystem(cfg.TestsRootDir)
}

func TestcaseLocalStorageServiceProvider(logger *slog.Logger, cfg *config.Config, repo repository.TestcaseLocalStorageRepository) (service.TestcaseLocalStorageService, error) {
	return service.NewTestcaseLocalStorageService(logger, repo, cfg.EnableCleanUp)
}
