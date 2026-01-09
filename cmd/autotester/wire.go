//go:build wireinject

package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared"
	sharedConfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/logger"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config, tracer trace.Tracer, isHeadless bool) (*gin.Engine, error) {
	wire.Build(
		shared.SharedProviderSet,
		LoggerProvider,
		OpenAiRepositoryProvider,
		OpenAiServiceProvider,
		FileSystemProvider,
		TestCaseParquetWrapperProvider,
		S3WrapperProvider,
		repository.NewTestcaseLocalStorageRepository,
		TestCaseStorageRepositoryProvider,
		service.NewValidatorService,
		service.NewTestcaseStorageService,
		TestcaseLocalStorageServiceProvider,
		RouterProvider,
		handler.NewAutotesterController,
		service.NewGeneratePromptService,
		TaglistConfigProvider,
		DockerClientProvider,
		service.NewDocker,
		service.NewChatManager,
		service.NewChatStorageService,
		repository.NewChatStorageRepository,
		ChatParquetConfigProvider,
		ChatParquetWrapperProvider,
		ChatSummaryParquetWrapperProvider,
		MetricsServiceProvider,
		DatabaseRepositoryProvider,
		service.NewAuthService,
	)

	return nil, nil
}

func RouterProvider(logger *slog.Logger, controller *handler.AutotesterController, isHeadless bool) (*gin.Engine, error) {
	return application.NewRouter(logger, controller, isHeadless)
}

func TaglistConfigProvider(cfg *config.Config) *sharedConfig.Taglist {
	return cfg.TaglistConfig
}

// LoggerProvider provides a new logger.
func LoggerProvider(cfg *config.Config) *slog.Logger {
	return logger.NewLogger(cfg.LogLevel)
}

// OpenAiRepositoryProvider provides a new OpenAI repository.
func OpenAiRepositoryProvider(client sharedRepo.OpenAIClient, cfg *config.Config, tracer trace.Tracer) (sharedRepo.OpenAI, error) {
	return sharedRepo.NewOpenAiRepository(client, cfg.Timeout, tracer)
}

// OpenAiServiceProvider provides a new OpenAI service.
func OpenAiServiceProvider(repo sharedRepo.OpenAI, tracer trace.Tracer) (sharedService.OpenAI, error) {
	return sharedService.NewOpenAI(repo, tracer)
}

// ChatParquetWrapperProvider provides a new session summary parquet wrapper.
func ChatParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig, tracer trace.Tracer) (wrapperService.ParquetFileWrapper[entity.Chat], error) {
	return wrapperService.NewParquetWrapper[entity.Chat](logger, cfg, tracer)
}

// ChatSummaryParquetWrapperProvider provides a new ChatSummary parquet wrapper.
func ChatSummaryParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig, tracer trace.Tracer) (wrapperService.ParquetFileWrapper[entity.ChatSummary], error) {
	return wrapperService.NewParquetWrapper[entity.ChatSummary](logger, cfg, tracer)
}

// TestCaseParquetWrapperProvider provides a new test case parquet wrapper with default parquet config.
func TestCaseParquetWrapperProvider(logger *slog.Logger, cfg wrapperEntity.ParquetConfig, tracer trace.Tracer) (wrapperService.ParquetFileWrapper[entity.TestCase], error) {
	return wrapperService.NewParquetWrapper[entity.TestCase](logger, cfg, tracer)
}

func S3WrapperProvider(logger *slog.Logger, cfg *config.Config, tracer trace.Tracer) (wrapperService.S3StorageWrapper, error) {
	config := wrapperEntity.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapperService.NewS3Wrapper(logger, config, tracer)
}

// FileSystemProvider provides a new filesystem.
func FileSystemProvider(cfg *config.Config) (repository.FileSystem, error) {
	return repository.NewOSFileSystem(cfg.TestsRootDir)
}

func DockerClientProvider() (service.DockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func TestcaseLocalStorageServiceProvider(logger *slog.Logger, cfg *config.Config, repo repository.TestcaseLocalStorageRepository) (service.TestcaseLocalStorageService, error) {
	return service.NewTestcaseLocalStorageService(logger, repo, cfg.EnableCleanUp)
}

func TestCaseStorageRepositoryProvider(
	logger *slog.Logger,
	s3Wrapper wrapperService.S3StorageWrapper,
	parquetWrapper wrapperService.ParquetFileWrapper[entity.TestCase],
	cfg *config.Config,
	tracer trace.Tracer,
) (repository.TestcaseStorageRepository, error) {
	return repository.NewTestcaseStorageRepository(logger, s3Wrapper, parquetWrapper, cfg.S3TestcasePrefix, tracer)
}

// ChatParquetConfigProvider provides a ParquetConfig for chat parquet files.
func ChatParquetConfigProvider() wrapperEntity.ParquetConfig {
	return wrapperService.DefaultParquetConfig()
}

func MetricsServiceProvider(logger *slog.Logger) (sharedService.MetricsService, error) {
	return sharedService.NewMetricsService("autotester", logger)
}

func DatabaseRepositoryProvider() (repository.TokenDatabase, error) {
	godotenv.Load(".env.dev")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("no db_url set in .env.dev")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %s", err)
	}
	return database.New(dbConn), nil
}
