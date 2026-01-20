package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterAPIRepository handles HTTP communication with the Autotester backend API.
type AutotesterAPIRepository interface {
	// GetTemplate fetches the test generation template from the API.
	GetTemplate(ctx context.Context) (*entity.TemplateResponse, error)

	// ValidatePrompt validates a prompt against the API.
	ValidatePrompt(ctx context.Context, request *entity.GenerateTestRequest) (*entity.ValidatePromptResponse, error)

	// GenerateTest sends a test generation request to the API.
	GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error)

	// SaveTest persists a generated test to the API.
	SaveTest(ctx context.Context, request *entity.SaveTestRequest) (*entity.SaveTestResponse, error)

	// RunTest executes a test by ID via the API.
	RunTest(ctx context.Context, request *entity.RunTestRequest) (*entity.RunTestResponse, error)
}

type autotesterAPIRepository struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    string
	tracer     trace.Tracer
}

// NewAutotesterAPIRepository creates a new instance of AutotesterAPIRepository.
// It initializes the repository with an HTTP client, base URL, and logger for API communication.
func NewAutotesterAPIRepository(logger *slog.Logger, httpClient *http.Client, baseURL string, tracer trace.Tracer) (AutotesterAPIRepository, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &autotesterAPIRepository{
		logger:     logger,
		httpClient: httpClient,
		baseURL:    baseURL,
		tracer:     tracer,
	}, nil
}

func (a *autotesterAPIRepository) GetTemplate(ctx context.Context) (*entity.TemplateResponse, error) {
	url := fmt.Sprintf("%s/api/v1/template", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.GetTemplate")
	defer span.End()

	req, err := a.newJSONRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		return nil, err
	}

	var result entity.TemplateResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		return nil, err
	}

	a.logger.Debug("Successfully fetched template", "templateLength", len(result.Content))
	return &result, nil
}

func (a *autotesterAPIRepository) ValidatePrompt(ctx context.Context, request *entity.GenerateTestRequest) (*entity.ValidatePromptResponse, error) {
	url := fmt.Sprintf("%s/api/v1/validate", a.baseURL)
	req, err := a.newJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		a.logger.Error("Failed to validate prompt", "error", err)
		return nil, err
	}

	var result entity.ValidatePromptResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		return nil, err
	}

	a.logger.Debug("Successfully validated prompt")
	return &result, nil
}

func (a *autotesterAPIRepository) GenerateTest(ctx context.Context, request *entity.GenerateTestRequest) (*entity.GenerateTestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/chat", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.GenerateTest")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", request.UserId),
		attribute.String("conversation.id", request.ConversationId),
	)

	req, err := a.newJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		return nil, err
	}

	var result entity.GenerateTestResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		return nil, err
	}

	a.logger.Debug("Successfully generated test")
	return &result, nil
}

func (a *autotesterAPIRepository) SaveTest(ctx context.Context, request *entity.SaveTestRequest) (*entity.SaveTestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/saveLocal", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.SaveTest")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", request.UserId),
		attribute.String("conversation.id", request.ConversationId),
	)

	req, err := a.newJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		return nil, err
	}

	var result entity.SaveTestResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		return nil, err
	}

	a.logger.Debug("Successfully saved test locally")
	return &result, nil
}

func (a *autotesterAPIRepository) RunTest(ctx context.Context, request *entity.RunTestRequest) (*entity.RunTestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/run", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.RunTest")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", request.TestId),
		attribute.String("user.id", request.UserId),
		attribute.String("conversation.id", request.ConversationId),
	)

	req, err := a.newJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		return nil, err
	}

	var result entity.RunTestResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		return nil, err
	}

	a.logger.Debug("Successfullywie d test run")
	return &result, nil
}

// newJSONRequest creates an HTTP request with an optional JSON body and sets
// the Content-Type header when a body is provided.
// --- helpers bound to the repository struct ---
func (a *autotesterAPIRepository) newJSONRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// doAndDecode executes the request, checks for a 200 OK response, and decodes
// the JSON response into result (if non-nil). It logs errors using the provided logger.
func doAndDecode[T any](client *http.Client, logger *slog.Logger, req *http.Request, result *T) error {
	if err := assert.NotNil(client, logger, req); err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to execute request", "error", err, "url", req.URL.String())
		return fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read error response body", "error", err, "status", resp.StatusCode)
			return fmt.Errorf("unexpected status code: %d (failed to read body: %w)", resp.StatusCode, err)
		}
		logger.Error("Unexpected status code", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			logger.Error("Failed to decode response", "error", err)
			return fmt.Errorf("failed to decode response: %w", err)
		}
	} else {
		// Drain any remaining response body so the HTTP connection can be reused.
		// Check and propagate any error instead of ignoring it.
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			logger.Error("Failed to discard response body", "error", err)
			return fmt.Errorf("failed to read response body: %w", err)
		}
	}

	return nil
}
