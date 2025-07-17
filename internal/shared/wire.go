package shared

import (
	"github.com/google/wire"
	"github.com/sashabaranov/go-openai"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// SharedProviderSet provides a set of providers that is shared between all domains.
// nolint:gochecknoglobals
var SharedProviderSet = wire.NewSet(
	OpenAiClientProvider,
	service.NewOpenAIService,
)

// OpenAiClientProvider provides a new OpenAI client.
func OpenAiClientProvider() (repository.OpenAIClient, error) {
	return openai.NewClient(build.OpenAIKey), nil
}
