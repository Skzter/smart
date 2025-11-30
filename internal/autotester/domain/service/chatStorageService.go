package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ErrChatNotFound is returned when no chat exists for the given userId/chatId.
var ErrChatNotFound = repository.ErrChatNotFound

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, summary *entity.Chat) error
	// LoadChat retrieves a Chat object from storage by a key generated from the provided userId and chatId.
	LoadChat(ctx context.Context, userId string, chatId string) (*entity.Chat, error)
	// FindByUserId retrieves an all ChatSummarys associated with the given userId
	// The resulting slice is ordered by updatedAt in descending order
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

// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
func (s *chatStorageService) SaveChat(ctx context.Context, summary *entity.Chat) error {
	if err := assert.NotNil(summary); err != nil {
		return err
	}
	if err := summary.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, summary)
}

// LoadChat retrieves a Chat object from storage by a key generated from the provided userId and chatId.
func (s *chatStorageService) LoadChat(ctx context.Context, userId string, chatId string) (*entity.Chat, error) {
	if err := assert.StringNotEmpty(userId); err != nil {
		return nil, fmt.Errorf("userId must not be empty")
	}
	if err := assert.StringNotEmpty(chatId); err != nil {
		return nil, fmt.Errorf("chatId must not be empty")
	}

	chat, err := s.repo.Read(ctx, userId, chatId)
	if err != nil {
		if errors.Is(err, repository.ErrChatNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	if err := chat.Validate(); err != nil {
		return nil, fmt.Errorf("retrieved invalid chat from s3: %w", err)
	}

	return chat, nil
}

// FindByUserId retrieves an all ChatSummarys associated with the given userId
// The resulting slice is ordered by updatedAt in descending order
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
