package repository

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageRepository defines the interface for a repository managing Chat entities.
type ChatStorageRepository interface {
	// Create stores the provided Chat object, as well as a generated ChatSummary Object in the underlying storage system.
	// The Storage key is generated from the entites userId and chatId, so duplicate entities will be overwriten
	Create(ctx context.Context, obj *entity.Chat) error

	// Read retrieves a Chat object from storage by a key generated from the provided userId and chatId.
	Read(ctx context.Context, chatId string) (*entity.Chat, error)

	// Delete removes linked Chat and ChatSummary objects from storage by their chatId and userId.
	Delete(ctx context.Context, chatId string) error

	// FindByUserId retrieves an slice of all ChatSummarys associated with the given userId
	ListAll(ctx context.Context) ([]*entity.ChatSummary, error)
}

const prefixChat = "chat"

// chatStorageRepository implements the ChatStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type chatStorageRepository struct {
	s3Wrapper             service.S3StorageWrapper
	chatParquetWrapper    service.ParquetFileWrapper[entity.Chat]
	summaryParquetWrapper service.ParquetFileWrapper[entity.ChatSummary]
	logger                *slog.Logger
	tracer                trace.Tracer
}

// NewChatStorageRepository creates a new repository for Chat entities.
// Returns the repository or an error.
func NewChatStorageRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	chatParquetWrapper service.ParquetFileWrapper[entity.Chat],
	summaryParquetWrapper service.ParquetFileWrapper[entity.ChatSummary],
	tracer trace.Tracer,
) (ChatStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, chatParquetWrapper, summaryParquetWrapper, tracer); err != nil {
		return nil, err
	}
	return &chatStorageRepository{
		s3Wrapper:             s3Wrapper,
		chatParquetWrapper:    chatParquetWrapper,
		summaryParquetWrapper: summaryParquetWrapper,
		logger:                logger,
		tracer:                tracer,
	}, nil
}

// Create stores the provided Chat object, as well as a generated ChatSummary Object in the underlying storage system.
// The Storage key is generated from the entites userId and chatId, so duplicate entities will be overwriten
// Returns an error if unsuccessfull, or nil otherwise
// nolint:dupl
func (r *chatStorageRepository) Create(ctx context.Context, obj *entity.Chat) error {
	if err := assert.NotNil(ctx, obj); err != nil {
		return err
	}

	ctx, span := r.tracer.Start(ctx, "chatStorageRepository.Create")
	defer span.End()
	span.SetAttributes(
		attribute.String("chat.id", obj.Id),
		attribute.String("chat.author", obj.Author),
		attribute.Int("chat.message_count", len(obj.Messages)),
	)

	summary := entity.ChatSummary{
		ChatId:         obj.Id,
		Author:         obj.Author,
		Groups:         obj.Groups,
		LastModifiedBy: obj.LastModifiedBy,
		Title:          obj.Title,
		CreatedAt:      obj.CreatedAt,
		UpdatedAt:      obj.UpdatedAt,
	}

	chatParquet, err := r.chatParquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize chat object")
		return err
	}

	summaryParquet, err := r.summaryParquetWrapper.WriteStructToParquet(ctx, summary)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize chat summary")
		return err
	}

	chatkey, summaryKey := generateKeys(obj.Id)
	err = r.s3Wrapper.UploadParquetFile(ctx, chatkey, chatParquet, map[string]string{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload chat parquet")
		return err
	}

	if err := r.s3Wrapper.UploadParquetFile(ctx, summaryKey, summaryParquet, map[string]string{}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload chat summary parquet")
		if err := r.s3Wrapper.DeleteParquetFile(ctx, chatkey); err != nil {
			r.logger.Error("WARNING: chat uploaded without summary, removal not possible", "key", chatkey, "err", err)
		}
		return err
	}

	r.logger.Debug("create: chat successfully written and uploaded",
		slog.String("key", chatkey),
		slog.String("type", "chat"),
	)

	r.logger.Debug("create: chatSummary successfully written and uploaded",
		slog.String("key", summaryKey),
		slog.String("type", "chatSummary"),
	)

	span.AddEvent("chat and summary stored", trace.WithAttributes(
		attribute.String("chat.key", chatkey),
		attribute.String("summary.key", summaryKey),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Read retrieves a Chat object from storage by a key generated from the provided userId and chatId.
// Returns the Chat or an error.
// nolint:dupl
func (r *chatStorageRepository) Read(ctx context.Context, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	key, _ := generateKeys(chatId)

	ctx, span := r.tracer.Start(ctx, "chatStorageRepository.Read")
	defer span.End()
	span.SetAttributes(
		attribute.String("chat.id", chatId),
	)

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download chat parquet")
		return nil, err
	}

	items, err := r.chatParquetWrapper.ReadStructsFromParquet(ctx, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read chat parquet")
		return nil, err
	}
	if len(items) == 0 {
		err := sharedErrors.ErrChatNotFound
		span.RecordError(err)
		span.SetStatus(codes.Error, "no chat found for user and chat id")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return &items[0], nil
}

// Delete removes linked Chat and ChatSummary objects from storage by their chatId and userId.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *chatStorageRepository) Delete(ctx context.Context, chatId string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	chatkey, summaryKey := generateKeys(chatId)

	ctx, span := r.tracer.Start(ctx, "chatStorageRepository.Delete")
	defer span.End()
	span.SetAttributes(
		attribute.String("chat.id", chatId),
	)

	if err := r.s3Wrapper.DeleteParquetFile(ctx, chatkey); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete chat parquet")
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", chatkey),
	)

	if err := r.s3Wrapper.DeleteParquetFile(ctx, summaryKey); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete chat summary parquet")
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", summaryKey),
	)

	span.AddEvent("chat and summary deleted", trace.WithAttributes(
		attribute.String("chat.key", chatkey),
		attribute.String("summary.key", summaryKey),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// ListAll retrieves all ChatSummarys
// Returns a slice of ChatSummary Objects or an error
func (r *chatStorageRepository) ListAll(ctx context.Context) ([]*entity.ChatSummary, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	ctx, span := r.tracer.Start(ctx, "chatStorageRepository.ListAll")
	defer span.End()

	keys, err := r.s3Wrapper.ListParquetFiles(ctx, fmt.Sprintf("%s/summary", prefixChat))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list summary parquet files")
		return nil, fmt.Errorf("failed to list chat summary parquet files: %w", err)
	}
	if len(keys) == 0 {
		err := fmt.Errorf("no chat summary files found in storage")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no chat summaries found")
		return nil, err
	}
	result := make([]*entity.ChatSummary, 0, len(keys))
	for _, key := range keys {
		parquetData, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Error("ListAll: failed to download file", "key", key, "error", err)
			continue
		}
		summary, err := r.summaryParquetWrapper.ReadStructsFromParquet(ctx, parquetData)
		if err != nil {
			r.logger.Error("ListAll: failed to read parquet data", "key", key, "error", err)
			continue
		}
		if len(summary) != 1 {
			r.logger.Error("ListAll: incorrect number of structs in parquet file", "key", key, "number", len(summary))
			continue
		}
		result = append(result, &summary[0])
	}

	r.logger.Debug("ListAll: finished loading chat summaries", slog.Int("count", len(result)))
	span.AddEvent("chat summaries loaded", trace.WithAttributes(
		attribute.Int("summary.count", len(result)),
	))
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func generateKeys(chatId string) (string, string) {
	return fmt.Sprintf("%s/full/%s.parquet", prefixChat, chatId),
		fmt.Sprintf("%s/summary/%s.parquet", prefixChat, chatId)
}
