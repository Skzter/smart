package repository

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageRepository defines the interface for a repository managing Chat entities.
type ChatStorageRepository interface {
	// Create stores a new Chat object in the underlying storage system.
	Create(ctx context.Context, obj *entity.Chat) error

	// Read retrieves a Chat object from storage by its key.
	Read(ctx context.Context, userId string, chatId string) (*entity.Chat, error)

	// Update overwrites an existing Chat object in storage.
	Update(ctx context.Context, obj *entity.Chat) error

	// Delete removes a Chat object from storage by its id and userId.
	Delete(ctx context.Context, userId string, chatId string) error

	FindByUserID(ctx context.Context, userId string) ([]*entity.ChatSummary, error)
}

const prefixChat = "chat"

// chatStorageRepository implements the ChatStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type chatStorageRepository struct {
	s3Wrapper             service.S3StorageWrapper
	chatParquetWrapper    service.ParquetFileWrapper[entity.Chat]
	summaryParquetWrapper service.ParquetFileWrapper[entity.ChatSummary]
	logger                *slog.Logger
}

// NewChatStorageRepository creates a new repository for Chat entities.
// Returns the repository or an error.
// nolint:lll
func NewChatStorageRepository(logger *slog.Logger, s3Wrapper service.S3StorageWrapper, chatParquetWrapper service.ParquetFileWrapper[entity.Chat], summaryParquetWrapper service.ParquetFileWrapper[entity.ChatSummary]) (ChatStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, chatParquetWrapper, summaryParquetWrapper); err != nil {
		return nil, err
	}
	return &chatStorageRepository{
		s3Wrapper:             s3Wrapper,
		chatParquetWrapper:    chatParquetWrapper,
		summaryParquetWrapper: summaryParquetWrapper,
		logger:                logger,
	}, nil
}

// Create serializes the given Chat object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *chatStorageRepository) Create(ctx context.Context, obj *entity.Chat) error {
	if err := assert.NotNil(ctx, obj); err != nil {
		return err
	}

	if err := obj.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	summary := entity.ChatSummary{
		ChatId:       obj.Id,
		UserId:       obj.UserId,
		Title:        obj.Title,
		CreatedAt:    obj.CreatedAt,
		UpdatedAt:    obj.UpdatedAt,
		MessageCount: len(obj.Messages),
	}

	chatParquet, err := r.chatParquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		return err
	}

	summaryParquet, err := r.summaryParquetWrapper.WriteStructToParquet(summary)
	if err != nil {
		return err
	}

	chatkey, summaryKey := generateKeys(obj.UserId, obj.Id)
	err = r.s3Wrapper.UploadParquetFile(ctx, chatkey, chatParquet, map[string]string{})
	if err != nil {
		return err
	}

	if err := r.s3Wrapper.UploadParquetFile(ctx, summaryKey, summaryParquet, map[string]string{}); err != nil {
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

	return nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first Chat found.
// Returns the Chat or an error.
// nolint:dupl
func (r *chatStorageRepository) Read(ctx context.Context, userId string, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	key, _ := generateKeys(userId, chatId)

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		return nil, err
	}

	items, err := r.chatParquetWrapper.ReadStructsFromParquet(data)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no data found for key %s", key)
	}
	if err := (&items[0]).Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided Chat.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *chatStorageRepository) Update(ctx context.Context, obj *entity.Chat) error {
	if err := assert.NotNil(ctx, obj); err != nil {
		return err
	}

	if err := obj.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	chatkey, _ := generateKeys(obj.UserId, obj.Id)
	exists, err := r.s3Wrapper.FileExists(ctx, chatkey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no files exist for session %s, user %s with key: %s", obj.Id, obj.UserId, chatkey)
	}
	return r.Create(ctx, obj)
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *chatStorageRepository) Delete(ctx context.Context, userId string, chatId string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	chatkey, summaryKey := generateKeys(userId, chatId)

	if err := r.s3Wrapper.DeleteParquetFile(ctx, chatkey); err != nil {
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", chatkey),
	)

	if err := r.s3Wrapper.DeleteParquetFile(ctx, summaryKey); err != nil {
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", summaryKey),
	)

	return nil
}

// List Metadata of all parquest files with given prefix.
func (r *chatStorageRepository) FindByUserID(ctx context.Context, userId string) ([]*entity.ChatSummary, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	keys, err := r.s3Wrapper.ListParquetFiles(ctx, fmt.Sprintf("%s/%s/summary", prefixChat, userId))
	if err != nil {
		return nil, fmt.Errorf("failed to list session summary parquet files: %w", err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no session summary files found in storage")
	}
	result := make([]*entity.ChatSummary, 0, len(keys))
	for _, key := range keys {
		parquetData, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Error("ListAll: failed to download file", "key", key, "error", err)
			continue
		}
		summary, err := r.summaryParquetWrapper.ReadStructsFromParquet(parquetData)
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
	return result, nil
}

func generateKeys(userId string, chatId string) (string, string) {
	return fmt.Sprintf("%s/%s/full/%s.parquet", prefixChat, userId, chatId),
		fmt.Sprintf("%s/%s/summary/%s.parquet", prefixChat, userId, chatId)
}
