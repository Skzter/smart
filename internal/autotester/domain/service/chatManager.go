package service

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatManager defines the behaviour for loading and saving chats used by the autotester.
// Implementations must handle persistence and any required initialization.
type ChatManager interface {
	LoadChat(context.Context, entity.UserRequest) (*entity.Chat, error)
	SaveChat(context.Context, *entity.Chat, string) error
}

// chatManager is the concrete implementation of Chat that delegates persistence to a storage service
// and enriches chatManager records with timestamps and configured prompts.
type chatManager struct {
	storageService ChatStorageService
	logger         *slog.Logger
	cfg            config.Config
	tracer         trace.Tracer
}

// NewChatManager constructs a new Chat service instance. It validates required dependencies and
// copies the provided config into the service.
func NewChatManager(logger *slog.Logger, storageService ChatStorageService, cfg *config.Config, trace trace.Tracer) (ChatManager, error) {
	if err := assert.NotNil(logger, storageService, cfg, trace); err != nil {
		return nil, err
	}
	return &chatManager{
		storageService: storageService,
		logger:         logger,
		cfg:            *cfg,
		tracer:         trace,
	}, nil
}

// LoadChat either loads an existing chat from storage (when request.ChatId is set) or
// creates and returns a new chat object initialized with a generated UUID, creation time,
// empty message slice and the initial user prompt.
func (c *chatManager) LoadChat(ctx context.Context, request entity.UserRequest) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	ctx, span := c.tracer.Start(ctx, "chatManager.LoadChat")
	defer span.End()

	if request.ChatId == "" {
		chat := entity.NewChat(request.UserId, []*entity.Message{})
		c.logger.Info("creating new chat", "user", request.UserId, "id", chat.Id)
		span.AddEvent("new chat created", trace.WithAttributes(
			attribute.String("user", request.UserId),
			attribute.String("id", chat.Id),
		))

		span.SetStatus(codes.Ok, "")
		return chat, nil
	}

	chat, err := c.storageService.LoadChat(ctx, request.ChatId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while loading chat")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return chat, nil
}

// SaveChat updates timestamps and the last-used prompts on the chat before delegating
// persistence to the storage service.
func (c *chatManager) SaveChat(ctx context.Context, chat *entity.Chat, userId string) error {
	if err := assert.NotNil(ctx, chat); err != nil {
		return err
	}
	ctx, span := c.tracer.Start(ctx, "chatManager.SaveChat")
	defer span.End()

	chat.UpdatedAt = time.Now().UTC()
	chat.LastModifiedBy = userId

	if err := c.storageService.SaveChat(ctx, chat); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat")
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}
