package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	// ReadTestLogStream opens a connection to the backend SSE stream.
	// This method is blocking and writes events to the provided channel.
	// Returns an error if the stream cannot be established or fails.
	ReadTestLogStream(ctx context.Context, testId string, eventsCh chan<- *entity.LogEvent) error
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
	if err := assert.NotNil(logger, httpClient); err != nil {
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode")
		return nil, err
	}

	a.logger.Debug("Successfully fetched template", "templateLength", len(result.Content))
	span.SetStatus(codes.Ok, "")
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
		attribute.String("chat.Id", request.ChatId),
	)

	req, err := a.newJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		a.logger.Error("Failed to create request", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create request")
		return nil, err
	}

	var result entity.GenerateTestResponse
	if err := doAndDecode(a.httpClient, a.logger, req, &result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode")
		return nil, err
	}

	a.logger.Debug("Successfully generated test")
	span.SetStatus(codes.Ok, "")
	return &result, nil
}

func (a *autotesterAPIRepository) SaveTest(ctx context.Context, request *entity.SaveTestRequest) (*entity.SaveTestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/saveLocal", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.SaveTest")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", request.UserId),
		attribute.String("conversation.id", request.ChatId),
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
	span.SetStatus(codes.Ok, "")
	return &result, nil
}

func (a *autotesterAPIRepository) RunTest(ctx context.Context, request *entity.RunTestRequest) (*entity.RunTestResponse, error) {
	url := fmt.Sprintf("%s/api/v1/run", a.baseURL)

	ctx, span := a.tracer.Start(ctx, "autotesterAPIRepository.RunTest")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", request.TestId),
		attribute.String("user.id", request.UserId),
		attribute.String("conversation.id", request.ChatId),
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

	a.logger.Debug("Successfully run test")
	return &result, nil
}

// ReadTestLogStream establishes an SSE connection to the backend stream and
// parses events in a line-oriented manner.
// The method is blocking and follows the lifecycle of `ctx`
func (a *autotesterAPIRepository) ReadTestLogStream(ctx context.Context, testId string, eventsCh chan<- *entity.LogEvent) error {
	url := fmt.Sprintf("%s/api/v1/test/%s/stream", a.baseURL, testId)
	a.logger.Debug("Establishing SSE stream to backend", "testId", testId, "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to stream: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	var currentEvent entity.LogEvent
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		line = strings.TrimRight(line, "\r\n")
		a.logger.Debug("Read line from stream", "testId", testId, "line", line)

		if strings.HasPrefix(line, ":") {
			continue
		}

		if line == "" {
			if currentEvent.Event != "" || currentEvent.Data != "" {
				eventCopy := currentEvent
				select {
				case <-ctx.Done():
					return ctx.Err()
				case eventsCh <- &eventCopy:
					a.logger.Debug("Successfully sent SSE event to channel", "testId", testId, "event", eventCopy.Event)
				}
				currentEvent = entity.LogEvent{}
			}
			continue
		}

		if after, ok := strings.CutPrefix(line, "event:"); ok {
			currentEvent.Event = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "data:"); ok {
			dataPart := strings.TrimSpace(after)
			if currentEvent.Data != "" {
				currentEvent.Data += "\n" + dataPart
			} else {
				currentEvent.Data = dataPart
			}
		}
	}

	return nil
}

// newJSONRequest creates an HTTP request with an optional JSON body and sets
// the Content-Type header when a body is provided.
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
