package OpenAIRepository

import (
	"context"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

type RepositoryInterface interface {
	CreateRequest(entity.Request, context.Context) entity.Response
}
