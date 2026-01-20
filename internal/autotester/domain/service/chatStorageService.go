package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, summary *entity.Chat) error
	// LoadChat retrieves a Chat object from storage by a key generated from the provided userId and chatId.
	LoadChat(ctx context.Context, chatId string) (*entity.Chat, error)
	// FindByUserId retrieves an all ChatSummarys associated with the given userId
	// The resulting slice is ordered by updatedAt in descending order
	LoadSummaries(ctx context.Context, groupIds ...string) ([]*entity.ChatSummary, error)
}

// chatStorageService implements the ChatStorageService interface
// and provides logic for storing Chat entities via the underlying repository.
type chatStorageService struct {
	logger    *slog.Logger
	repo      repository.ChatStorageRepository
	validator Validator
	cache     Cache
	tracer    trace.Tracer
	metrics   sharedService.MetricsService
}

// NewChatStorageService creates a new ChatStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewChatStorageService(
	logger *slog.Logger,
	repo repository.ChatStorageRepository,
	validator Validator,
	cache Cache,
	tracer trace.Tracer,
	metrics sharedService.MetricsService,
) (ChatStorageService, error) {
	if err := assert.NotNil(logger, repo, validator, metrics, cache); err != nil {
		return nil, err
	}

	return &chatStorageService{
		logger:    logger,
		repo:      repo,
		validator: validator,
		cache:     cache,
		tracer:    tracer,
		metrics:   metrics,
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "error during validation")
		return err
	}
	if err := s.repo.Create(ctx, chat); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat")
		return err
	}

	if err := s.cache.Store(ctx, chat); err != nil {
		s.logger.Warn("cache store error", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat in cache")
	}
	return nil
}

// LoadChat retrieves a Chat object from storage by a key generated from the provided chatId.
func (s *chatStorageService) LoadChat(ctx context.Context, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.LoadChat")
	defer span.End()

	if err := assert.StringNotEmpty(chatId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing chatId")
		return nil, fmt.Errorf("chatId must not be empty")
	}

	start := time.Now()

	// cache miss produces nil, nil
	cachedChat, err := s.cache.LookUp(ctx, chatId)
	if err != nil {
		s.logger.Info("cache lookup error", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache access error")
	}

	if cachedChat != nil {
		if err := s.validator.ValidateChat(ctx, cachedChat); err != nil {
			s.logger.Warn("retrieved invalid chat from cache", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "retrieved invalid chat from cache")
		} else {
			s.metrics.IncCacheHit()
			s.metrics.RecordCacheDuration(time.Since(start), "hit")
			return cachedChat, nil
		}
	}

	chat, err := s.repo.Read(ctx, chatId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while reading chat")
		return nil, err
	}

	if err := s.validator.ValidateChat(ctx, chat); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "retrieved invalid chat")
		return nil, fmt.Errorf("retrieved invalid chat from s3: %w", err)
	}

	if err := s.cache.Store(ctx, chat); err != nil {
		s.logger.Warn("cache store error", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat in cache")
	}

	s.metrics.IncCacheMiss()
	s.metrics.RecordCacheDuration(time.Since(start), "miss")
	span.SetStatus(codes.Ok, "")
	return chat, nil
}

// LoadSummaries retrieves an all ChatSummarys associated with any of the given groupIds
// The resulting slice is ordered by updatedAt in descending order
func (s *chatStorageService) LoadSummaries(ctx context.Context, groupIds ...string) ([]*entity.ChatSummary, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.LoadSummaries")
	defer span.End()

	if err := assert.StringsNotEmpty(groupIds...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid groupId")
		return nil, fmt.Errorf("groupId must not be empty string")
	}
	summaries, err := s.repo.ListAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while retrieving chatSummaries")
		return nil, err
	}

	if len(groupIds) > 0 {
		summaries = slices.DeleteFunc(summaries, func(s *entity.ChatSummary) bool {
			return !slices.ContainsFunc(s.Groups, func(id string) bool {
				return slices.Contains(groupIds, id)
			})
		})
	}

	// sort in descending order by UpdatedAt
	slices.SortFunc(summaries, func(a *entity.ChatSummary, b *entity.ChatSummary) int {
		return -a.UpdatedAt.Compare(b.UpdatedAt)
	})
	span.SetStatus(codes.Ok, "")
	return summaries, nil
}
