package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ChatStorageService provides an interface to persist Chat entities.
type ChatStorageService interface {
	// SaveChat persists the provided Chat entity, as well as a generated ChatSummary entity into the storage.
	// Returns an error if the operation fails.
	SaveChat(ctx context.Context, chat *entity.Chat) error
	// LoadChat retrieves a Chat object from storage by a key generated from the provided chatId.
	LoadChat(ctx context.Context, chatId string) (*entity.Chat, error)
	// LoadSummaries retrieves all ChatSummarys associated with any of the given groupIds.
	// The resulting slice is ordered by updatedAt in descending order.
	// Returns whether more summaries exist.
	LoadSummaries(ctx context.Context, offset int, limit int, groupIds ...string) ([]*entity.ChatSummary, bool, error)
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

	lock      sync.RWMutex
	summaries []*entity.ChatSummary
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

	summaries, err := repo.ListAll(context.Background())
	if err != nil {
		return nil, err
	}

	slices.SortFunc(summaries, func(a *entity.ChatSummary, b *entity.ChatSummary) int {
		return a.Cmp(b)
	})

	return &chatStorageService{
		logger:    logger,
		repo:      repo,
		validator: validator,
		cache:     cache,
		tracer:    tracer,
		summaries: summaries,
		lock:      sync.RWMutex{},
		metrics:   metrics,
	}, nil
}

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

	summary := &entity.ChatSummary{
		ChatId:         chat.Id,
		Author:         chat.Author,
		Groups:         chat.Groups,
		LastModifiedBy: chat.LastModifiedBy,
		Title:          chat.Title,
		CreatedAt:      chat.CreatedAt,
		UpdatedAt:      chat.UpdatedAt,
	}

	if err := s.repo.Create(ctx, chat, summary); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat")
		return errors.ErrInternalServer
	}

	if err := s.cache.Store(ctx, chat); err != nil {
		s.logger.Debug("cache store error", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat in cache")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	insertPos, _ := sort.Find(len(s.summaries), func(i int) int {
		return summary.Cmp(s.summaries[i])
	})

	if index := slices.IndexFunc(s.summaries, func(existing *entity.ChatSummary) bool {
		return existing.ChatId == chat.Id
	}); index != -1 {
		copy(s.summaries[insertPos+1:index+1], s.summaries[insertPos:index])
		s.summaries[insertPos] = summary
	} else {
		s.summaries = slices.Insert(s.summaries, insertPos, summary)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *chatStorageService) LoadChat(ctx context.Context, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "chatStorageService.LoadChat")
	defer span.End()

	if err := assert.StringNotEmpty(chatId); err != nil {
		s.logger.Error("missing chatId", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing chatId")
		return nil, errors.ErrValidation
	}

	start := time.Now()

	// cache miss produces nil, nil
	cachedChat, err := s.cache.LookUp(ctx, chatId)
	if err != nil {
		s.logger.Debug("cache lookup error", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache access error")
	}

	if cachedChat != nil {
		if err := s.validator.ValidateChat(ctx, cachedChat); err != nil {
			s.logger.Debug("retrieved invalid chat from cache", "error", err)
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

	if err := s.cache.Store(ctx, chat); err != nil {
		s.logger.Debug("cache store error", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat in cache")
	}

	s.metrics.IncCacheMiss()
	s.metrics.RecordCacheDuration(time.Since(start), "miss")
	span.SetStatus(codes.Ok, "")
	return chat, nil
}

func (s *chatStorageService) LoadSummaries(ctx context.Context, offset int, limit int, groupIds ...string) ([]*entity.ChatSummary, bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, false, err
	}

	_, span := s.tracer.Start(ctx, "chatStorageService.LoadSummaries")
	defer span.End()

	if err := assert.StringsNotEmpty(groupIds...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid groupId")
		return nil, false, fmt.Errorf("groupId must not be empty string")
	}

	if err := assert.NumberGreaterThan(limit, 0); err != nil {
		return nil, false, err
	}

	if err := assert.NumberGreaterOrEqualThan(offset, 0); err != nil {
		return nil, false, err
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.summaries) == 0 {
		return s.summaries, false, nil
	}

	filteredSummaries := s.summaries
	if len(groupIds) > 0 {
		maxNeeded := offset + limit + 1
		filteredSummaries = findFromGroups(s.summaries, groupIds, maxNeeded)
	}

	if err := assert.NumberLessThan(offset, len(filteredSummaries)); err != nil {
		return nil, false, err
	}

	paginatedSummaries := filteredSummaries[offset:]

	// Apply limit and determine if more results exist
	hasMore := len(paginatedSummaries) > limit
	if hasMore {
		paginatedSummaries = paginatedSummaries[:limit]
	}

	span.SetStatus(codes.Ok, "")
	return paginatedSummaries, hasMore, nil
}

func findFromGroups(summaries []*entity.ChatSummary, groupIds []string, maxResults int) []*entity.ChatSummary {
	if maxResults <= 0 {
		return []*entity.ChatSummary{}
	}

	groupMap := make(map[string]bool, len(groupIds))
	for _, id := range groupIds {
		groupMap[id] = true
	}

	result := make([]*entity.ChatSummary, 0, maxResults)
	for _, summary := range summaries {
		if len(result) >= maxResults {
			break
		}

		for _, groupId := range summary.Groups {
			if groupMap[groupId] {
				result = append(result, summary)
				break
			}
		}
	}
	return result
}
