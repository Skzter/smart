//go:build wireinject

package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared"
	sharedConfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	wconfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/logger"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// InitializeApp initializes the application.
func InitializeApp(cfg *config.Config, tracer trace.Tracer) (*gin.Engine, error) {
	wire.Build(
		LoggerProvider,
		application.NewRouter,
		handler.NewSuproxyController,
		service.NewValidator,
		HTTPClientProvider,
		shared.SharedProviderSet,
		OpenAiRepositoryProvider,
		OpenAiServiceProvider,
		service.NewDatabaseService,
		DatabaseRepositoryProvider,
		service.NewTaglistSync,
		DatabaseParquetWrapperProvider,
		S3WrapperProvider,
		TagsearchServiceProvider,
		TaglistConfigProvider,
		MetricsServiceProvider,
		RedisCacheProvider,
		CacheServiceProvider,
	)

	return nil, nil
}

// MetricsServiceProvider provides a new metrics service.
func MetricsServiceProvider(logger *slog.Logger) (sharedService.MetricsService, error) {
	return sharedService.NewMetricsService("suproxy", logger)
}

// TaglistConfigProvider provides a new TaglistConfig.
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

func HTTPClientProvider() *http.Client {
	return &http.Client{}
}

func DatabaseParquetWrapperProvider(logger *slog.Logger, tracer trace.Tracer) (wrapper.ParquetFileWrapper[entity.DatabaseEntry], error) {
	return wrapper.NewParquetWrapper[entity.DatabaseEntry](logger, wrapper.DefaultParquetConfig(), tracer)
}

func S3WrapperProvider(logger *slog.Logger, cfg *config.Config, tracer trace.Tracer) (wrapper.S3StorageWrapper, error) {
	config := wconfig.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapper.NewS3Wrapper(logger, config, tracer)
}

// DatabaseRepositoryProvider provides a new DatabaseRepository.
func DatabaseRepositoryProvider(
	logger *slog.Logger,
	cfg *config.Config,
	s3wrapper wrapper.S3StorageWrapper,
	parquetWrapper wrapper.ParquetFileWrapper[entity.DatabaseEntry],
	tracer trace.Tracer,
) (repository.DatabaseRepository, error) {
	return repository.NewDatabaseRepository(
		logger,
		s3wrapper,
		parquetWrapper,
		tracer,
		cfg.EntryPrefix,
	)
}

func TagsearchServiceProvider(cfg *config.Config, s3 wrapper.S3StorageWrapper) (service.TagSearchService, error) {
	return service.NewTagSearchService(cfg, s3)
}

// RedisCacheProvider provides a new RedisCache
func RedisCacheProvider(log *slog.Logger, cfg *config.Config) (repository.Cache, error) {
	return repository.NewRedisCache(log, cfg)
}

// CacheServiceProvider provides a new CacheService
func CacheServiceProvider(log *slog.Logger, cfg *config.Config, repo repository.Cache) (service.CacheService, error) {
	return service.NewCacheService(log, cfg, repo)
}
