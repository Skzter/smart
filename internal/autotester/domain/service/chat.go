package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Chat defines the behaviour for loading and saving chat sessions used by the autotester.
// Implementations must handle persistence and any required initialization.
type Chat interface {
	LoadChat(context.Context, entity.UserRequest) (*entity.Chat, error)
	SaveChat(context.Context, *entity.Chat) error
}

// chat is the concrete implementation of Chat that delegates persistence to a storage service
// and enriches chat records with timestamps and configured prompts.
type chat struct {
	storageService ChatStorageService
	logger         *slog.Logger
	cfg            config.Config
}

// NewChat constructs a new Chat service instance. It validates required dependencies and
// copies the provided config into the service.
func NewChat(logger *slog.Logger, storageService ChatStorageService, cfg *config.Config) (Chat, error) {
	if err := assert.NotNil(logger, storageService, cfg); err != nil {
		return nil, err
	}
	return &chat{
		storageService: storageService,
		logger:         logger,
		cfg:            *cfg,
	}, nil
}

// LoadChat either loads an existing chat from storage (when request.ChatId is set) or
// creates and returns a new chat object initialized with a generated UUID, creation time,
// empty message slice and the initial user prompt.
func (c *chat) LoadChat(ctx context.Context, request entity.UserRequest) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	if request.ChatId == "" {
		id := uuid.NewString()
		c.logger.Info("creating new chat", "user", request.UserId, "id", id)
		return &entity.Chat{
			Id:        id,
			UserId:    request.UserId,
			CreatedAt: time.Now().UTC(),

			Messages: []sharedEntity.Message{},

			InitialUserPrompt: request.Prompt,
		}, nil
	}
	return c.storageService.LoadChat(ctx, request.UserId, request.ChatId)
}

// SaveChat updates timestamps and the last-used prompts on the chat before delegating
// persistence to the storage service.
func (c *chat) SaveChat(ctx context.Context, chat *entity.Chat) error {
	if err := assert.NotNil(ctx, chat); err != nil {
		return err
	}
	chat.UpdatedAt = time.Now().UTC()
	chat.LastValidationPrompt = c.cfg.Prompts.ValidationPrompt
	chat.LastAutoPlaywrightPrompt = c.cfg.Prompts.AutoPlaywrightPromptT

	return c.storageService.SaveChat(ctx, chat)
}
