//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared"
	sharedConfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	wconfig "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
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
func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		LoggerProvider,
		application.NewRouter,
		handler.NewSuproxyController,
		service.NewValidator,
		HTTPClientProvider,
		shared.SharedProviderSet,
		OpenAiRepositoryProvider,
		service.NewDatabaseService,
		DatabaseRepositoryProvider,
		service.NewTaglistSync,
		DatabaseParquetWrapperProvider,
		S3WrapperProvider,
		TagsearchServiceProvider,
		TaglistConfigProvider,
		RedisCacheProvider,
		CacheServiceProvider,
	)

	return nil, nil
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
func OpenAiRepositoryProvider(logger *slog.Logger, client sharedRepo.OpenAIClient, cfg *config.Config) (sharedRepo.OpenAI, error) {
	return sharedRepo.NewOpenAiRepository(logger, client, cfg.Timeout)
}

// HTTPClientProvider provides a new HTTP Client.
func HTTPClientProvider() *http.Client {
	return &http.Client{}
}

// DatabaseParquetWrapperProvider provides a new DatabaseParwuetWrapper.
func DatabaseParquetWrapperProvider(logger *slog.Logger) (wrapper.ParquetFileWrapper[entity.DatabaseEntry], error) {
	return wrapper.NewParquetWrapper[entity.DatabaseEntry](logger, wrapper.DefaultParquetConfig())
}

// S3WrapperProvider provides a new S3Wrapper.
func S3WrapperProvider(logger *slog.Logger, cfg *config.Config) (wrapper.S3StorageWrapper, error) {
	config := wconfig.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapper.NewS3Wrapper(logger, config)
}

// DatabaseRepositoryProvider provides a new DatabaseRepository.
func DatabaseRepositoryProvider(
	logger *slog.Logger,
	cfg *config.Config,
	s3wrapper wrapper.S3StorageWrapper,
	parquetWrapper wrapper.ParquetFileWrapper[entity.DatabaseEntry],
) (repository.DatabaseRepository, error) {
	return repository.NewDatabaseRepository(
		logger,
		s3wrapper,
		parquetWrapper,
		cfg.EntryPrefix,
	)
}

// TagsearchServiceProvider provides a new TagsearchService
func TagsearchServiceProvider(cfg *config.Config, s3 wrapper.S3StorageWrapper) (service.TagSearchService, error) {
	return service.NewTagSearchService(cfg, s3)
}

// RedisCacheProvider provides a new RedisCache
func RedisCacheProvider(log *slog.Logger, cfg *config.Config) (repository.Cache, error) {
	return repository.NewRedisCache(log, cfg)
}

// CacheServiceProvider provides a new CacheService
func CacheServiceProvider(log *slog.Logger, cfg *config.Config, repo repository.Cache) service.CacheService {
	return service.NewCacheService(log, cfg, repo)
}
