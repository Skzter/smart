package handler

import (
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterController is the controller for autotesting requests.
// It encapsulates logging and access to the OpenAI service.
type AutotesterController struct {
	config             *config.Config
	logger             *slog.Logger
	validationService  service.Validator
	generationService  service.GeneratePrompt
	saveLocalService   service.TestcaseLocalStorageService
	dockerService      service.Docker
	chatStorageService service.ChatStorageService
	chatManager        service.ChatManager
}

// NewAutotesterController creates a new AutotesterController.
// Returns an initialized controller or an error.
func NewAutotesterController(
	logger *slog.Logger,
	config *config.Config,
	validationService service.Validator,
	generationService service.GeneratePrompt,
	saveLocalService service.TestcaseLocalStorageService,
	dockerService service.Docker,
	chatStorageService service.ChatStorageService,
	chatManager service.ChatManager,
) (*AutotesterController, error) {
	if err := assert.NotNil(logger, config, validationService, generationService, saveLocalService, dockerService, chatStorageService, chatManager); err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:             logger,
		config:             config,
		validationService:  validationService,
		generationService:  generationService,
		saveLocalService:   saveLocalService,
		dockerService:      dockerService,
		chatStorageService: chatStorageService,
		chatManager:        chatManager,
	}, nil
}
