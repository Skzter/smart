package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/repository"
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

	return &entity.ExecuteTestResponse{Result: combined}, nil
}
