package repository

import (
	"context"
	"log/slog"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// RepositoryInterface defines methods for interacting with OpenAI API.
type RepositoryInterface interface {
	// CreateRequest sends a request to OpenAI API and returns the response.
	// It takes a Request entity containing the model and prompts,
	// a context for cancellation, and a logger for error reporting.
	CreateRequest(entity.Request, context.Context, *slog.Logger) (entity.Response, error)
}
