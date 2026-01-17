package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/trace"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

type responseUpdateService struct {
	logger           *slog.Logger
	tracer           trace.Tracer
	tagSearchService TagSearchService
	databaseRepo     repository.DatabaseRepository
}

// RespondeUpdateService defines the contract for updating persisted responses based on a given test request.
type RespondeUpdateService interface {
	UpdateResponse(ctx context.Context, mockRequest *entity.Request) error
}

// NewResponseUpdateService constructs a ResponseUpdateService with all required dependencies for updating persisted responses.
func NewResponseUpdateService(
	logger *slog.Logger,
	tracer trace.Tracer,
	tagSearchService TagSearchService,
	databaseRepo repository.DatabaseRepository,
) (*responseUpdateService, error) {
	if err := assert.NotNil(logger, tracer, tagSearchService, databaseRepo); err != nil {
		return nil, fmt.Errorf("dependency cannot be nil, %w", err)
	}

	return &responseUpdateService{
		logger:           logger,
		tracer:           tracer,
		tagSearchService: tagSearchService,
		databaseRepo:     databaseRepo,
	}, nil
}

// UpdateResponse locates a persisted response by tags, updates its offer data, and stores the modified response back to persistence.
func (s *responseUpdateService) UpdateResponse(
	ctx context.Context,
	mockRequest *entity.Request,
) error {
	if err := assert.NotNil(ctx, mockRequest); err != nil {
		return fmt.Errorf("invalid input, %w", err)
	}

	ctx, span := s.tracer.Start(ctx, "ResponseUpdateService.UpdateResponse")
	defer span.End()

	if mockRequest.Tags == "" {
		return fmt.Errorf("update response requires non-empty tags")
	}

	parquetFiles, err := s.tagSearchService.FindKeysByTags(ctx, mockRequest.Tags)
	if err != nil {
		return fmt.Errorf("tag-based search failed: %w", err)
	}

	if len(parquetFiles) == 0 {
		return fmt.Errorf("no parquet files found for tags: %s", mockRequest.Tags)
	}

	requestTags := splitRequestTags(mockRequest.Tags)

	dbEntry, err := s.selectBestMatchingResponse(ctx, parquetFiles, requestTags)
	if err != nil {
		return err
	}

	oldResponse := dbEntry.Response
	chunkedResponse, err := chunkOffers(oldResponse.Response)
	if err != nil {
		return fmt.Errorf("failed to chunk response, %w", err)
	}

	updatedChunks, err := updateChunks(chunkedResponse)
	if err != nil {
		return fmt.Errorf("failed to update chunks, %w", err)
	}

	updatedResponse, err := reassembleResponse(updatedChunks)
	if err != nil {
		return fmt.Errorf("reassembling of chunks into updated response failed, %w", err)
	}

	if err := s.databaseRepo.CreateRequest(ctx, *updatedResponse); err != nil {
		return fmt.Errorf("failed to save updated Response to s3, %w", err)
	}

	return nil
}

// selectBestMatchingResponse selects the persisted response with the highest tag overlap.
func (s *responseUpdateService) selectBestMatchingResponse(
	ctx context.Context,
	keys []string,
	requestTags []string,
) (*entity.DatabaseEntry, error) {
	var (
		bestEntry *entity.DatabaseEntry
		bestScore = -1
	)

	for _, key := range keys {
		entry, err := s.databaseRepo.ReadRequest(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to read parquet file %s: %w", key, err)
		}

		responseTags := splitResponseTags(entry.Tags)
		score := tagScore(requestTags, responseTags)

		if score > bestScore {
			bestScore = score
			bestEntry = entry
		}
	}

	if bestEntry == nil || bestScore <= 0 {
		return nil, fmt.Errorf("no suitable response found for tags: %v", requestTags)
	}

	return bestEntry, nil
}

// splitRequestTags parses comma-separated request tags.
func splitRequestTags(tags string) []string {
	raw := strings.Split(tags, ",")
	result := make([]string, 0, len(raw))

	for _, t := range raw {
		if trimmed := strings.TrimSpace(t); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// splitResponseTags extracts comparable tag values from a TagList.
func splitResponseTags(tags *sharedEntity.TagList) []string {
	if tags == nil {
		return nil
	}

	result := make([]string, 0, len(tags.Tags))
	for _, t := range tags.Tags {
		if t.Name != "" {
			result = append(result, t.Name)
		}
	}

	return result
}

// tagScore calculates the overlap between request and response tags.
func tagScore(requestTags, responseTags []string) int {
	set := make(map[string]struct{}, len(responseTags))
	for _, t := range responseTags {
		set[t] = struct{}{}
	}

	score := 0
	for _, t := range requestTags {
		if _, ok := set[t]; ok {
			score++
		}
	}

	return score
}

// chunkOffers parses a raw response JSON string and extracts offer data into an UpdateResponse structure for processing.
func chunkOffers(response string) (*entity.UpdateResponse, error) {
	// TODO: parse response JSON and extract data.items
	return nil, nil
}

// updateChunks applies request-dependent updates to offers and recalculates aggregated fields within an UpdateResponse.
func updateChunks(oldChunks *entity.UpdateResponse) (*entity.UpdateResponse, error) {
	// TODO: update offers and recalc aggregates
	return nil, nil
}

// reassembleResponse merges updated offer data back into the original response structure and returns a persistable database entry.
func reassembleResponse(newChunks *entity.UpdateResponse) (*entity.DatabaseEntry, error) {
	// TODO: reassembling implementation
	return nil, nil
}
