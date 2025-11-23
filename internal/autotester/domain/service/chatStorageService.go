package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, summary *entity.Chat) error
	LoadChat(ctx context.Context, chatId string, userId string) (*entity.Chat, error)
	UpdateChat(ct context.Context, chat *entity.Chat) error
	LoadUserChats(ctx context.Context, userID string) ([]*entity.ChatSummary, error)
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
	return s.repo.Create(ctx, summary)
}

func (s *chatStorageService) LoadChat(ctx context.Context, userId string, chatId string) (*entity.Chat, error) {
	if err := assert.StringNotEmpty(userId); err != nil {
		return nil, fmt.Errorf("userId must not be empty")
	}
	if err := assert.StringNotEmpty(chatId); err != nil {
		return nil, fmt.Errorf("chatId must not be empty")
	}

	return s.repo.Read(ctx, userId, chatId)
}

func (s *chatStorageService) UpdateChat(ctx context.Context, chat *entity.Chat) error {
	return s.repo.Update(ctx, chat)
}

func (s *chatStorageService) LoadUserChats(ctx context.Context, userId string) ([]*entity.ChatSummary, error) {
	if err := assert.StringNotEmpty(userId); err != nil {
		return nil, fmt.Errorf("key must not be empty")
	}
	summaries, err := s.repo.FindByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	// sort in descending order by UpdatedAt
	slices.SortFunc(summaries, func(a *entity.ChatSummary, b *entity.ChatSummary) int {
		return -a.UpdatedAt.Compare(b.UpdatedAt)
	})
	return summaries, nil
}
