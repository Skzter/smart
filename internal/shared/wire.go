package shared

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// SharedProviderSet provides a set of providers that is shared between all domains.
// nolint:gochecknoglobals
var SharedProviderSet = wire.NewSet(
	OpenAiClientProvider,
	TaglistStorageProvider,
	TagListParquetWrapperProvider,
	sharedService.NewTaglistStorage,
	RedisCacheProvider,
)

// OpenAiClientProvider provides a new OpenAI client.
func OpenAiClientProvider() (repository.OpenAIClient, error) {
	return openai.NewClient(build.OpenAIKey), nil
}

// TaglistStorageProvider provides a Tagliststorage repository
func TaglistStorageProvider(
	logger *slog.Logger,
	cfg *config.TaglistConfig,
	parquet wrapper.ParquetFileWrapper[entity.TagList],
	tracer trace.Tracer,
) (repository.TaglistStorage, error) {
	s3, err := S3WrapperProvider(logger, cfg, tracer)
	if err != nil {
		panic(err)
	}
	return repository.NewTaglistStorage(
		logger,
		s3,
		parquet,
		cfg.EntryPrefix,
		tracer,
	)
}

// S3WrapperProvider provides an S3Wrapper configured with the shared config
func S3WrapperProvider(logger *slog.Logger, cfg *config.TaglistConfig, tracer trace.Tracer) (wrapper.S3StorageWrapper, error) {
	config := wrapperEntity.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapper.NewS3Wrapper(logger, config, tracer)
}

// TagListParquetWrapperProvider provides a ParquetWrapper for Taglist
func TagListParquetWrapperProvider(logger *slog.Logger, tracer trace.Tracer) (wrapper.ParquetFileWrapper[entity.TagList], error) {
	return wrapper.NewParquetWrapper[entity.TagList](logger, wrapper.DefaultParquetConfig(), tracer)
}

// RedisCacheProvider provides a new RedisCache
func RedisCacheProvider(log *slog.Logger, cfg *config.RedisConfig, tracer trace.Tracer) (repository.Cache, error) {
	return repository.NewRedisCache(log, cfg, tracer)
}
