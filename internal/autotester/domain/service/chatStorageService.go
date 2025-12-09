package service

import (
	"context"
	"log/slog"
	"slices"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, summary *entity.Chat) error
	// LoadChat retrieves a Chat object from storage by a key generated from the provided userId and chatId.
	LoadChat(ctx context.Context, chatId string, userId string) (*entity.Chat, error)
	// FindByUserId retrieves an all ChatSummarys associated with the given userId
	// The resulting slice is ordered by updatedAt in descending order
	LoadUserChats(ctx context.Context, userID string) ([]*entity.ChatSummary, error)
}

// chatStorageService implements the ChatStorageService interface
// and provides logic for storing Chat entities via the underlying repository.
type chatStorageService struct {
	logger    *slog.Logger
	repo      repository.ChatStorageRepository
	validator Validator
	tracer    trace.Tracer
}

// NewChatStorageService creates a new ChatStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewChatStorageService(logger *slog.Logger, repo repository.ChatStorageRepository, validator Validator, tracer trace.Tracer) (ChatStorageService, error) {
	if err := assert.NotNil(logger, repo, validator); err != nil {
		return nil, err
	}

	return &chatStorageService{
		logger:    logger,
		repo:      repo,
		validator: validator,
		tracer:    tracer,
	}, nil
}

// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
func (s *chatStorageService) SaveChat(ctx context.Context, chat *entity.Chat) error {
	if err := assert.NotNil(ctx, chat); err != nil {
		return err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.SaveChat")
	defer span.End()

	if err := s.validator.ValidateChat(ctx, chat); err != nil {
		s.logger.Error("chat validation failed", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "error during validation")
		return errors.ErrValidation
	}
	if err := s.repo.Create(ctx, chat); err != nil {
		s.logger.Error("error while storing chat", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat")
		return errors.ErrInternalServer
	}
	return nil
}

// LoadChat retrieves a Chat object from storage by a key generated from the provided userId and chatId.
func (s *chatStorageService) LoadChat(ctx context.Context, userId string, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.LoadChat")
	defer span.End()

	if err := assert.StringNotEmpty(userId); err != nil {
		s.logger.Error("missing userId", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing userId")
		return nil, errors.ErrValidation
	}
	if err := assert.StringNotEmpty(chatId); err != nil {
		s.logger.Error("missing chatId", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing chatId")
		return nil, errors.ErrValidation
	}

	chat, err := s.repo.Read(ctx, userId, chatId)
	if err != nil {
		s.logger.Error("error while reading chat", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while reading chat")
		return nil, errors.ErrInternalServer
	}

	if err := s.validator.ValidateChat(ctx, chat); err != nil {
		s.logger.Error("retrieved invalid chat from storage", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "retrieved invalid chat")
		return nil, errors.ErrValidation
	}

	span.SetStatus(codes.Ok, "")
	return chat, nil
}

// FindByUserId retrieves an all ChatSummarys associated with the given userId
// The resulting slice is ordered by updatedAt in descending order
func (s *chatStorageService) LoadUserChats(ctx context.Context, userId string) ([]*entity.ChatSummary, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.LoadUserChats")
	defer span.End()

	if err := assert.StringNotEmpty(userId); err != nil {
		s.logger.Error("missing userId", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing userId")
		return nil, errors.ErrValidation
	}
	summaries, err := s.repo.FindByUserID(ctx, userId)
	if err != nil {
		s.logger.Error("error while retrieving chatSummaries", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while retrieving chatSummaries")
		return nil, errors.ErrInternalServer
	}

	// sort in descending order by UpdatedAt
	slices.SortFunc(summaries, func(a *entity.ChatSummary, b *entity.ChatSummary) int {
		return -a.UpdatedAt.Compare(b.UpdatedAt)
	})
	span.SetStatus(codes.Ok, "")
	return summaries, nil
}
