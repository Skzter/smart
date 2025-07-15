package shared

import (
	"github.com/google/wire"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// SharedProviderSet provides a set of providers that is shared between all domains.
// nolint:gochecknoglobals
var SharedProviderSet = wire.NewSet(
	OpenAiClientProvider,
	service.NewOpenAIService,
)

// OpenAiClientProvider provides a new OpenAI client.
func OpenAiClientProvider() (openai.Client, error) {
	return openai.NewClient(
		option.WithAPIKey(build.OpenAIKey),
	), nil
}
