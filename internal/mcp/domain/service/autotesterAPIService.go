package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterAPIService provides business logic for interacting with the Autotester API.
type AutotesterAPIService interface {
	// GetTemplate retrieves the test generation template.
	GetTemplate(ctx context.Context) (*entity.TemplateResponse, error)

	// GenerateTest creates a new test from the provided specification.
	GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error)

	// ExecuteTest runs an existing test by ID.
	ExecuteTest(ctx context.Context, request *entity.ExecuteTestRequest) (*entity.ExecuteTestResponse, error)
}

type autotesterAPIService struct {
	logger *slog.Logger
	repo   repository.AutotesterAPIRepository
}

// NewAutotesterAPIService creates a new service for the Autotester API.
// Expects a logger and a repository, checks both for nil.
func NewAutotesterAPIService(logger *slog.Logger, repo repository.AutotesterAPIRepository) (AutotesterAPIService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &autotesterAPIService{
		logger: logger,
		repo:   repo,
	}, nil
}

func (s *autotesterAPIService) GetTemplate(ctx context.Context) (*entity.TemplateResponse, error) {
	s.logger.Info("Fetching test template from API")

	template, err := s.repo.GetTemplate(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch template", "error", err)
		return nil, err
	}

	s.logger.Info("Successfully retrieved template", "templateLength", len(template.Content))
	return template, nil
}

func (s *autotesterAPIService) GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error) {
	s.logger.Info("Generating test via API")

	if err := assert.NotNil(request); err != nil {
		s.logger.Error("Invalid generate test request", "error", err)
		return nil, err
	}

	resp, err := s.repo.GenerateTest(ctx, request)
	if err != nil {
		s.logger.Error("Failed to generate test", "error", err)
		return nil, err
	}

	s.logger.Info("Successfully generated test", "resultLength", len(resp.Result))
	return resp, nil
}

func (s *autotesterAPIService) ExecuteTest(ctx context.Context, request *entity.ExecuteTestRequest) (*entity.ExecuteTestResponse, error) {
	s.logger.Info("Executing test via API")

	if err := assert.NotNil(request); err != nil {
		s.logger.Error("Invalid execute test request", "error", err)
		return nil, err
	}

	saveReq := &entity.SaveTestRequest{
		Code:           request.Test,
		UserId:         request.UserId,
		ConversationId: request.ConversationId,
	}

	saveResp, err := s.repo.SaveTest(ctx, saveReq)
	if err != nil {
		s.logger.Error("Failed to save test before execution", "error", err)
		return nil, err
	}

	runReq := &entity.RunTestRequest{
		TestId:         saveResp.TestId,
		UserId:         request.UserId,
		ConversationId: request.ConversationId,
	}

	runResp, err := s.repo.RunTest(ctx, runReq)
	if err != nil {
		s.logger.Error("Failed to run test", "error", err)
		return nil, err
	}

	combined := fmt.Sprintf("saved:testId=%s action=%s; runResult=%s", saveResp.TestId, saveResp.Action, runResp.Result)

	s.logger.Info("Successfully executed test", "summary", combined)

	return &entity.ExecuteTestResponse{Result: combined}, nil
}
