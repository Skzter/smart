package repository

import (
	"context"
	"log/slog"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

type RepositoryInterface interface {
	CreateRequest(entity.Request, context.Context, *slog.Logger) (entity.Response, error)
}
