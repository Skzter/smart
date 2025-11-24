//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/docker/docker/client"
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
	sharedConfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
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
		LogFileSystemProvider,
		repository.NewTestcaseLocalStorageRepository,
		service.NewValidatorService,
		TestcaseLocalStorageServiceProvider,
		application.NewRouter,
		handler.NewAutotesterController,
		service.NewGeneratePromptService,
		TaglistConfigProvider,
		DockerClientProvider,
		service.NewDocker,
		shared.S3WrapperProvider,
		ChatParquetWrapperProvider,
		ChatSummaryParquetWrapperProvider,
		repository.NewChatStorageRepository,
		service.NewChatStorageService,
	)

	return nil, nil
}

func TaglistConfigProvider(cfg *config.Config) *sharedConfig.Taglist {
	return cfg.TaglistConfig
}

// LoggerProvider provides a new logger.
func LoggerProvider(cfg *config.Config) *slog.Logger {
	return logger.NewLogger(cfg.LogLevel)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiRepositoryProvider(client sharedRepo.OpenAIClient, cfg *config.Config) (sharedRepo.OpenAI, error) {
	return sharedRepo.NewOpenAiRepository(client, cfg.Timeout)
}

// ChatParquetWrapperProvider provides a new session summary parquet wrapper.
func ChatParquetWrapperProvider(logger *slog.Logger) (wrapperService.ParquetFileWrapper[entity.Chat], error) {
	return wrapperService.NewParquetWrapper[entity.Chat](logger, wrapperService.DefaultParquetConfig())
}

// ChatParquetWrapperProvider provides a new chat summary parquet wrapper.
func ChatSummaryParquetWrapperProvider(logger *slog.Logger) (wrapperService.ParquetFileWrapper[entity.ChatSummary], error) {
	return wrapperService.NewParquetWrapper[entity.ChatSummary](logger, wrapperService.DefaultParquetConfig())
}

// TestCaseParquetWrapperProvider provides a new test case parquet wrapper.
func TestCaseParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig) (wrapperService.ParquetFileWrapper[entity.TestCase], error) {
	return wrapperService.NewParquetWrapper[entity.TestCase](logger, cfg)
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

// FileSystemProvider provides a new filesystem.
func FileSystemProvider(cfg *config.Config) (repository.FileSystem, error) {
	return repository.NewOSFileSystem(cfg.TestsRootDir)
}

// LogFileSystemProvider provides a filesystem for logs
func LogFileSystemProvider(cfg *config.Config) (repository.LogFileSystem, error) {
	return repository.NewLogFileSystem(cfg.LogDirAutopw)
}

func DockerClientProvider() (service.DockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func TestcaseLocalStorageServiceProvider(logger *slog.Logger, cfg *config.Config, repo repository.TestcaseLocalStorageRepository) (service.TestcaseLocalStorageService, error) {
	return service.NewTestcaseLocalStorageService(logger, repo, cfg.EnableCleanUp)
}
