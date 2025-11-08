package shared

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/sashabaranov/go-openai"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// SharedProviderSet provides a set of providers that is shared between all domains.
// nolint:gochecknoglobals
var SharedProviderSet = wire.NewSet(
	OpenAiClientProvider,
	service.NewOpenAI,
	service.NewTaglistStorage,
	TaglistStorageProvider,
)

// OpenAiClientProvider provides a new OpenAI client.
func OpenAiClientProvider() (repository.OpenAIClient, error) {
	return openai.NewClient(build.OpenAIKey), nil
}

// TaglistStorageProvider provides a Tagliststorage repository
func TaglistStorageProvider(
	logger *slog.Logger,
	cfg *config.Taglist,
) (repository.TaglistStorage, error) {
	s3, err := S3WrapperProvider(logger, cfg)
	if err != nil {
		panic(err)
	}
	parquet, err := TagListParquetWrapperProvider(logger)
	if err != nil {
		panic(err)
	}
	return repository.NewTaglistStorage(
		logger,
		s3,
		parquet,
		cfg.EntryPrefix)
}

// S3WrapperProvider provides an S3Wrapper configured with the shared config
func S3WrapperProvider(logger *slog.Logger, cfg *config.Taglist) (wrapper.S3StorageWrapper, error) {
	config := wrapperEntity.S3Config{
		Region:    cfg.Region,
		Bucket:    cfg.Bucket,
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}
	return wrapper.NewS3Wrapper(logger, config)
}

// TagListParquetWrapperProvider provides a ParquetWrapper for Taglist
func TagListParquetWrapperProvider(logger *slog.Logger) (wrapper.ParquetFileWrapper[entity.TagList], error) {
	return wrapper.NewParquetWrapper[entity.TagList](logger, wrapper.DefaultParquetConfig())
}
