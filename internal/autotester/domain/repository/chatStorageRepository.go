package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageRepository defines the interface for a repository managing Chat entities.
type ChatStorageRepository interface {
	// Create stores a new Chat object in the underlying storage system.
	Create(ctx context.Context, obj *entity.Chat) error

	// Read retrieves a Chat object from storage by its key.
	Read(ctx context.Context, key string) (*entity.Chat, error)

	// Update overwrites an existing Chat object in storage.
	Update(ctx context.Context, obj *entity.Chat, key string) error

	// Delete removes a Chat object from storage by its key.
	Delete(ctx context.Context, key string) error

	ListAll(ctx context.Context) ([]*entity.ChatSummary, error)
}

const prefixChat = "chat"

// chatStorageRepository implements the ChatStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type chatStorageRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.Chat]
	logger         *slog.Logger
}

// NewChatStorageRepository creates a new repository for Chat entities.
// Returns the repository or an error.
// nolint:lll
func NewChatStorageRepository(logger *slog.Logger, s3Wrapper service.S3StorageWrapper, parquetWrapper service.ParquetFileWrapper[entity.Chat]) (ChatStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}
	return &chatStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given Chat object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *chatStorageRepository) Create(ctx context.Context, obj *entity.Chat) error {
	if err := validateChat(obj); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		return err
	}

	now := fmt.Sprintf("%d", time.Now().UTC().Unix())
	metadata := map[string]string{
		"chat-id":       obj.Id,
		"user-id":       obj.UserId,
		"title":         obj.Title,
		"created-at":    now,
		"upated-at":     now,
		"message-count": fmt.Sprintf("%d", len(obj.Messages)),
	}

	key := generateChatKey()

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return err
	}

	r.logger.Debug("create: object successfully written and uploaded",
		slog.String("key", key),
		slog.String("type", "chat"),
	)

	return nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first Chat found.
// Returns the Chat or an error.
// nolint:dupl
func (r *chatStorageRepository) Read(ctx context.Context, key string) (*entity.Chat, error) {
	if err := assert.StringNotEmpty(key); err != nil {
		return nil, fmt.Errorf("key must not be empty")
	}

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		return nil, err
	}

	items, err := r.parquetWrapper.ReadStructsFromParquet(data)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no data found for key %s", key)
	}
	if err := validateChat(&items[0]); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided Chat.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *chatStorageRepository) Update(ctx context.Context, obj *entity.Chat, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := validateChat(obj); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	exists, err := r.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if key exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("cannot update: key does not exist")
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	metadata := map[string]string{}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload updated object: %w", err)
	}

	r.logger.Debug("update: object successfully overwritten",
		slog.String("key", key),
		slog.String("type", "chat"),
	)

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *chatStorageRepository) Delete(ctx context.Context, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", key),
	)
	return nil
}

// List Metadata of all parquest files with given prefix.
func (r *chatStorageRepository) ListAll(ctx context.Context) ([]*entity.ChatSummary, error) {
	keys, err := r.s3Wrapper.ListParquetFiles(ctx, fmt.Sprint(prefixChat+"/"))
	if err != nil {
		return nil, fmt.Errorf("failed to list session summary parquet files: %w", err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no session summary files found in storage")
	}
	result := make([]*entity.ChatSummary, 0, len(keys))

	for _, key := range keys {
		_, metadata, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Warn("ListAll: failed to download parquet file",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}

		createdSec, err := strconv.ParseInt(metadata["created-at"], 10, 64)
		if err != nil {
			r.logger.Warn("ListAll: file has corrupt metadata",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}

		updatedSec, err := strconv.ParseInt(metadata["updated-at"], 10, 64)
		if err != nil {
			r.logger.Warn("ListAll: file has corrupt metadata",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}

		mcount, err := strconv.Atoi(metadata["message-count"])
		if err != nil {
			r.logger.Warn("ListAll: file is missing metadata",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}

		summary := &entity.ChatSummary{
			ChatId:       metadata["chat-id"],
			UserId:       metadata["user-id"],
			Title:        metadata["title"],
			CreatedAt:    time.Unix(createdSec, 0),
			UpdatedAt:    time.Unix(updatedSec, 0),
			MessageCount: mcount,
		}
		result = append(result, summary)
	}

	r.logger.Debug("ListAll: finished loading chat summaries", slog.Int("count", len(result)))
	return result, nil
}

// generateChatKey creates a unique S3 key for a Chat object.
// The format is: "chat/chat_<timestamp>"
func generateChatKey() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%s_%d", prefixChat, prefixChat, timestamp)
}

// validateChat validates a Chat entity.
// Returns an error if any required field is empty or invalid.
func validateChat(chat *entity.Chat) error {
	switch {
	case chat == nil:
		return errors.New("pointer is nil")
	case chat.InitialPrompt == "":
		return errors.New("initial prompt is empty")
	case chat.SystemPrompt == "":
		return errors.New("system prompt is empty")
	case chat.Id == "":
		return errors.New("id is empty")
	case chat.UserId == "":
		return errors.New("userId is empty")
	case len(chat.Messages) == 0:
		return errors.New("contains no messages")
	}

	for _, msg := range chat.Messages {
		switch {
		case msg.Body == "":
			return errors.New("contains empty messages")
		case msg.Id == "":
			return errors.New("contains messages with empty id")
		case msg.Role == "":
			return errors.New("contains message with empty role")
		}
	}
	return nil
}
