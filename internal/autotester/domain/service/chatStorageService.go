package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, summary *entity.Chat) error
	LoadChat(ctx context.Context, key string) (*entity.Chat, error)
	ListSummaries(ctx context.Context) ([]*entity.ChatSummary, error)
}

// chatStorageService implements the ChatStorageService interface
// and provides logic for storing Chat entities via the underlying repository.
type chatStorageService struct {
	logger *slog.Logger
	repo   repository.ChatStorageRepository
}

// NewChatStorageService creates a new ChatStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewChatStorageService(logger *slog.Logger, repo repository.ChatStorageRepository) (ChatStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &chatStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveChat saves the given Chat entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (s *chatStorageService) SaveChat(ctx context.Context, summary *entity.Chat) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	return s.repo.Create(ctx, summary)
}

func (s *chatStorageService) LoadChat(ctx context.Context, key string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	return s.repo.Read(ctx, key)
}

func (s *chatStorageService) ListSummaries(ctx context.Context) ([]*entity.ChatSummary, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	return s.repo.ListAll(ctx)
}
