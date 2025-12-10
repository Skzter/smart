package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterAPIRepository handles HTTP communication with the Autotester backend API.
type AutotesterAPIRepository interface {
	// GetTemplate fetches the test generation template from the API.
	GetTemplate(ctx context.Context) (*entity.TemplateResponse, error)

	// GenerateTest sends a test generation request to the API.
	GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error)

	// SaveTest persists a generated test to the API.
	SaveTest(ctx context.Context, request *entity.SaveTestRequest) (*entity.SaveTestResponse, error)

	// RunTest executes a test by ID via the API.
	RunTest(ctx context.Context, requst *entity.RunTestRequest) (*entity.RunTestResponse, error)
}

type autotesterAPIRepository struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    string
}

// NewAutotesterAPIRepository creates a new instance of AutotesterAPIRepository.
// It initializes the repository with an HTTP client, base URL, and logger for API communication.
func NewAutotesterAPIRepository(logger *slog.Logger, httpClient *http.Client, baseURL string) (AutotesterAPIRepository, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &autotesterAPIRepository{
		logger:     logger,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

func (a *autotesterAPIRepository) GetTemplate(ctx context.Context) (*entity.TemplateResponse, error) {
	url := fmt.Sprintf("%s/api/v1/template", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.logger.Error("Failed to execute request", "error", err, "url", url)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			a.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			a.logger.Error("Failed to read error response body", "error", err, "status", resp.StatusCode)
			return nil, fmt.Errorf("unexpected status code: %d (failed to read body: %w)", resp.StatusCode, err)
		}
		a.logger.Error("Unexpected status code", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result entity.TemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		a.logger.Error("Failed to decode response", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	a.logger.Info("Successfully fetched template", "templateLength", len(result.Content))
	return &result, nil
}

func (a *autotesterAPIRepository) GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error) {
	// TODO: Implement
	// url := fmt.Sprintf("%s/api/v1/chat", a.baseURL)
	return nil, nil
}

func (a *autotesterAPIRepository) SaveTest(ctx context.Context, request *entity.SaveTestRequest) (*entity.SaveTestResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (a *autotesterAPIRepository) RunTest(ctx context.Context, requst *entity.RunTestRequest) (*entity.RunTestResponse, error) {
	// TODO: Implement
	return nil, nil
}
