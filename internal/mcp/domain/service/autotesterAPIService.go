package service

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

const jwtKey = "jwt"

// AutotesterAPIService provides business logic for interacting with the Autotester API.
type AutotesterAPIService interface {
	// GetTemplate retrieves the test generation template.
	GetTemplate(ctx context.Context) (*entity.TemplateResponse, error)

	// GenerateTest creates a new test from the provided specification after validation.
	GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestToolResponse, error)

	// ExecuteTest runs an existing test by ID.
	ExecuteTest(ctx context.Context, request *entity.ExecuteTestRequest) (*entity.ExecuteTestResponse, error)

	// ReadTestLogStream reads log events from the backend stream and stores them.
	// It marks the stream as complete when the backend stream is exhausted.
	ReadTestLogStream(ctx context.Context, testId string) error
}

type autotesterAPIService struct {
	logger *slog.Logger
	repo   repository.AutotesterAPIRepository
	store  store.TestLogStreamStore
}

// NewAutotesterAPIService creates a new service for the Autotester API.
// Expects a logger, a repository, and a test log stream store, checks all for nil.
func NewAutotesterAPIService(logger *slog.Logger, repo repository.AutotesterAPIRepository, store store.TestLogStreamStore) (AutotesterAPIService, error) {
	if err := assert.NotNil(logger, repo, store); err != nil {
		return nil, err
	}

	return &autotesterAPIService{
		logger: logger,
		repo:   repo,
		store:  store,
	}, nil
}

func (s *autotesterAPIService) GetTemplate(ctx context.Context) (*entity.TemplateResponse, error) {
	s.logger.Debug("Fetching test template from API")

	token, _ := ctx.Value(jwtKey).(string)

	template, err := s.repo.GetTemplate(ctx, token)
	if err != nil {
		s.logger.Error("Failed to fetch template", "error", err)
		return nil, err
	}

	s.logger.Info("Successfully retrieved template", "templateLength", len(template.Content))
	return template, nil
}

func (s *autotesterAPIService) GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestToolResponse, error) {
	s.logger.Debug("Generating test via API")

	if err := assert.NotNil(request); err != nil {
		s.logger.Error("Invalid generate test request", "error", err)
		return nil, err
	}

	token, _ := ctx.Value(jwtKey).(string)

	valid, err := s.repo.ValidatePrompt(ctx, request, token)
	if err != nil {
		s.logger.Error("Failed to validate prompt", "error", err)
		return nil, err
	}

	// If the validation result contains a message body, the prompt is considered incomplete or invalid.
	// In this case, we return early with the validation feedback.
	if valid.Result.Body != "" {
		s.logger.Info("Prompt validation failed", "validationMessage", valid.Result.Body)
		return &entity.GenerateTestToolResponse{
			ValidateMsg: &valid.Result,
			UserId:      valid.UserId,
			ChatId:      valid.ChatId,
		}, nil
	}
	s.logger.Debug("Prompt validation successful")

	request.ChatId = valid.ChatId

	resp, err := s.repo.GenerateTest(ctx, request, token)
	if err != nil {
		s.logger.Error("Failed to generate test", "error", err)
		return nil, err
	}

	s.logger.Info("Successfully generated test", "resultLength", len(resp.Result.Body))
	return &entity.GenerateTestToolResponse{
		GenerateMsg: &resp.Result,
		UserId:      resp.UserId,
		ChatId:      resp.ChatId,
	}, nil
}

func (s *autotesterAPIService) ExecuteTest(ctx context.Context, request *entity.ExecuteTestRequest) (*entity.ExecuteTestResponse, error) {
	s.logger.Debug("Executing test via API")

	if err := assert.NotNil(request); err != nil {
		s.logger.Error("Invalid execute test request", "error", err)
		return nil, err
	}

	token, _ := ctx.Value(jwtKey).(string)

	saveReq := &entity.SaveTestRequest{
		Code:   request.Test,
		UserId: request.UserId,
		ChatId: request.ChatId,
	}

	saveResp, err := s.repo.SaveTest(ctx, saveReq, token)
	if err != nil {
		s.logger.Error("Failed to save test before execution", "error", err)
		return nil, err
	}

	runReq := &entity.RunTestRequest{
		TestId: saveResp.TestId,
		UserId: request.UserId,
		ChatId: request.ChatId,
	}

	runResp, err := s.repo.RunTest(ctx, runReq, token)
	if err != nil {
		s.logger.Error("Failed to run test", "error", err)
		return nil, err
	}

	combined := fmt.Sprintf("saved:testId=%s action=%s; runResult=%s", saveResp.TestId, saveResp.Action, runResp.Result)

	s.logger.Info("Successfully executed test", "summary", combined)

	return &entity.ExecuteTestResponse{Result: combined, TestId: saveResp.TestId}, nil
}

// ReadTestLogStream reads SSE events and stores them.
//
// IMPORTANT: This method must be called at most once concurrently per testId.
// Calling this method multiple times simultaneously for the same testId can lead to race conditions
// when both streams write to the same store, resulting in data loss or inconsistency.
//
// The current architecture ensures this constraint at the call site, but callers must guarantee
// that a new ReadTestLogStream call for the same testId only happens after the previous stream
// has completed (indicated by calling store.CompleteStream).
//
// Technical: producer–consumer using two goroutines — producer reads SSE into a
// buffered channel, consumer drains the channel and writes to the store. Coordination
// is via `errgroup` and a derived context.
func (s *autotesterAPIService) ReadTestLogStream(ctx context.Context, testId string) error {
	s.logger.Info("Start reading and processing log stream", "testId", testId)
	const rawEventsCapacity = 2048
	rawEventsCh := make(chan *entity.LogEvent, rawEventsCapacity)
	wg, groupCtx := errgroup.WithContext(ctx)

	// PRODUCER: reads logstream
	wg.Go(func() error {
		defer close(rawEventsCh)
		s.logger.Debug("PRODUCER: Starting SSE stream read", "testId", testId)

		if err := s.repo.ReadTestLogStream(groupCtx, testId, rawEventsCh); err != nil {
			s.logger.Warn("PRODUCER: SSE stream ended with error", "testId", testId, "error", err)
			return err
		}
		s.logger.Debug("PRODUCER: SSE reader stopped", "testId", testId)
		return nil
	})

	// CONSUMER: add logEvents to store
	wg.Go(func() error {
		s.logger.Debug("CONSUMER: Starting event processor", "testId", testId)

		for {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case event, ok := <-rawEventsCh:
				if !ok {
					s.logger.Info("Stream read and processed", "testId", testId)
					return nil
				}
				if event != nil {
					s.logger.Debug("CONSUMER: Received event",
						"testId", testId,
						"event", event.Event,
					)
					s.store.AddEvent(testId, *event)
				}
			}
		}
	})

	// MONITOR: check channel capacity while waiting for goroutines to complete
	monitorDone := make(chan struct{})
	go func() {
		const capacityThreshold = rawEventsCapacity * 9 / 10
		lastWarned := false

		for {
			select {
			case <-monitorDone:
				return
			case <-groupCtx.Done():
				return
			default:
				currentLoad := len(rawEventsCh)
				if currentLoad >= capacityThreshold && !lastWarned {
					s.logger.Warn("Event channel at 90% capacity",
						"testId", testId,
						"currentLoad", currentLoad,
						"capacity", rawEventsCapacity,
					)
					lastWarned = true
				} else if currentLoad < capacityThreshold {
					lastWarned = false
				}
			}
		}
	}()

	err := wg.Wait()
	close(monitorDone)

	s.store.CompleteStream(testId)

	if err != nil {
		s.logger.Error("Log stream processing failed", "testId", testId, "error", err)
		return err
	}

	return nil
}
