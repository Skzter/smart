package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseStorageService provides an interface to persist TestCase entities.
type TestcaseStorageService interface {
	// SaveTestCase persists the provided TestCase entity into the storage.
	// Returns an error if the operation fails.
	SaveTestcase(ctx context.Context, testcase *entity.TestCase, userId string) (string, error)

	// ReadAllMetadataWithFilter retrieves all test case metadata and applies optional filters.
	// Filters by author, testcaseId, and creation timestamps. Returns filtered metadata or an error.
	ReadAllMetadataWithFilter(ctx context.Context, filter *entity.GetRemoteTestcaseRequest) ([]*entity.TestcaseMetadata, error)
}

// testcaseStorageService implements the TestcaseStorageService interface
// and provides logic for storing TestCase entities via the underlying repository.
type testcaseStorageService struct {
	logger *slog.Logger
	repo   repository.TestcaseStorageRepository
	tracer trace.Tracer
}

// NewTestcaseStorageService creates a new TestcaseStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewTestcaseStorageService(
	logger *slog.Logger,
	repo repository.TestcaseStorageRepository,
	tracer trace.Tracer,
) (TestcaseStorageService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, err
	}

	return &testcaseStorageService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

// SaveTestCase saves the given TestCase entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (t *testcaseStorageService) SaveTestcase(ctx context.Context, testcase *entity.TestCase, userId string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		t.logger.Error("context validation failed", "error", err)
		return "", err
	}

	ctx, span := t.tracer.Start(ctx, "testcaseStorageService.SaveTestCase")
	defer span.End()

	key, err := t.repo.Create(ctx, testcase, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save test case")
		return "", err
	}

	t.logger.Debug("testcase successfully saved",
		slog.String("testID", testcase.TestID),
	)
	span.SetStatus(codes.Ok, "")

	return key, nil
}

// ReadAllMetadataWithFilter retrieves all test case metadata and applies optional filters.
// Filters by author, testcaseId, and creation timestamps. Returns filtered metadata or an error.
func (t *testcaseStorageService) ReadAllMetadataWithFilter(ctx context.Context, filterParams *entity.GetRemoteTestcaseRequest) ([]*entity.TestcaseMetadata, error) {
	if err := assert.NotNil(ctx); err != nil {
		t.logger.Error("context validation failed", "error", err)
		return nil, err
	}

	ctx, span := t.tracer.Start(ctx, "testcaseStorageService.ReadAllMetadata")
	defer span.End()

	metadata, err := t.repo.ReadAllMetadata(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read all metadata")
		return nil, err
	}

	t.logger.Debug("all metadata successfully read",
		slog.Int("count", len(metadata)),
	)
	span.AddEvent("metadata successfully read", trace.WithAttributes(
		attribute.Int("count", len(metadata)),
	))

	filteredMetadata := t.filter(metadata, filterParams)

	t.logger.Debug("metadata filtered",
		slog.Int("original", len(metadata)),
		slog.Int("filtered", len(filteredMetadata)),
	)
	span.AddEvent("metadata filtered", trace.WithAttributes(
		attribute.Int("original", len(metadata)),
		attribute.Int("filtered", len(filteredMetadata)),
	))

	span.SetStatus(codes.Ok, "")
	return filteredMetadata, nil
}

func (t *testcaseStorageService) filter(metadata []*entity.TestcaseMetadata, filterParams *entity.GetRemoteTestcaseRequest) []*entity.TestcaseMetadata {
	result := make([]*entity.TestcaseMetadata, 0, len(metadata))

	for _, m := range metadata {
		if t.passesAllFilters(m, filterParams) {
			result = append(result, m)
		}
	}

	return result
}

func (t *testcaseStorageService) passesAllFilters(metadata *entity.TestcaseMetadata, filterParams *entity.GetRemoteTestcaseRequest) bool {
	if filterParams.Author != "" && metadata.Author != filterParams.Author {
		return false
	}

	if filterParams.TestcaseId != "" && !strings.Contains(metadata.Key, filterParams.TestcaseId) {
		return false
	}

	if filterParams.CreatedAfter != "" {
		afterTimestamp := t.iso8601ToUnix(filterParams.CreatedAfter)
		createdUnix := t.stringToInt64(metadata.Created)
		if afterTimestamp > 0 && createdUnix < afterTimestamp {
			return false
		}
	}

	if filterParams.CreatedBefore != "" {
		beforeTimestamp := t.iso8601ToUnix(filterParams.CreatedBefore)
		createdUnix := t.stringToInt64(metadata.Created)
		if beforeTimestamp > 0 && createdUnix > beforeTimestamp {
			return false
		}
	}

	return true
}

func (t *testcaseStorageService) stringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.logger.Warn("failed to parse unix timestamp string", "timestamp", s, "error", err)
		return 0
	}
	return val
}

func (t *testcaseStorageService) iso8601ToUnix(iso8601 string) int64 {
	parsedTime, err := time.Parse(time.RFC3339, iso8601)
	if err != nil {
		t.logger.Warn("failed to parse ISO 8601 timestamp", "timestamp", iso8601, "error", err)
		return 0
	}
	return parsedTime.UTC().Unix()
}
