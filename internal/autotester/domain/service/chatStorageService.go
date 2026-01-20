package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
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

func sort(ids []*entity.ChatSummary) {
	// sort in descending order by UpdatedAt
	slices.SortFunc(ids, func(a *entity.ChatSummary, b *entity.ChatSummary) int {
		if updated := -a.UpdatedAt.Compare(b.UpdatedAt); updated != 0 {
			return updated
		}
		return strings.Compare(a.ChatId, b.ChatId)
	})
}

// chatStorageService implements the ChatStorageService interface
// and provides logic for storing Chat entities via the underlying repository.
type chatStorageService struct {
	logger    *slog.Logger
	repo      repository.ChatStorageRepository
	validator Validator
	tracer    trace.Tracer

	summaries []*entity.ChatSummary
}

// NewChatStorageService creates a new ChatStorageService instance.
// Returns an error if required arguments are nil or if loading existing summaries fails.
func NewChatStorageService(logger *slog.Logger, repo repository.ChatStorageRepository, validator Validator, tracer trace.Tracer) (ChatStorageService, error) {
	if err := assert.NotNil(logger, repo, validator); err != nil {
		return nil, err
	}

	summaries, err := repo.ListAll(context.Background())
	if err != nil {
		return nil, err
	}

	sort(summaries)

	return &chatStorageService{
		logger:    logger,
		repo:      repo,
		validator: validator,
		tracer:    tracer,
		summaries: summaries,
	}, nil
}

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

	var summary *entity.ChatSummary
	if index := slices.IndexFunc(s.summaries, func(existing *entity.ChatSummary) bool {
		return existing.ChatId == chat.Id
	}); index != -1 {
		summary = s.summaries[index]
		summary.UpdatedAt = chat.UpdatedAt
		summary.Groups = chat.Groups
	} else {
		summary = &entity.ChatSummary{
			ChatId:         chat.Id,
			Author:         chat.Author,
			Groups:         chat.Groups,
			LastModifiedBy: chat.LastModifiedBy,
			Title:          chat.Title,
			CreatedAt:      chat.CreatedAt,
			UpdatedAt:      chat.UpdatedAt,
		}
		s.summaries = append(s.summaries, summary)
		sort(s.summaries)
	}

	if err := s.repo.Create(ctx, chat, summary); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing chat")
		return err
	}

	return nil
}

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
	result := make([]*entity.ChatSummary, 0, len(summaries))

	for _, summary := range summaries {
		if len(result) >= maxResults {
			break
		}

		belongsToGroup := slices.ContainsFunc(summary.Groups, func(groupId string) bool {
			return slices.Contains(groupIds, groupId)
		})

		if belongsToGroup {
			result = append(result, summary)
		}
	}

	return slices.Clip(result)
}
