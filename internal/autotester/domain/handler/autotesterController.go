package handler

import (
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterController is the controller for autotesting requests.
// It encapsulates logging and access to the OpenAI service.
type AutotesterController struct {
	config                       *config.Autotester
	logger                       *slog.Logger
	validationService            service.Validator
	generationService            service.GeneratePrompt
	localTestcaseStorageService  service.TestcaseLocalStorageService
	dockerService                service.Docker
	chatStorageService           service.ChatStorageService
	remoteTestcaseStorageService service.TestcaseStorageService
	mediaStorageService          service.MediaStorageService
	chatManager                  service.ChatManager
	groupManager                 service.GroupManager
	tracer                       trace.Tracer
	metricsService               shared.MetricsService
	authService                  service.Auth
}

// NewAutotesterController creates a new AutotesterController.
// Returns an initialized controller or an error.
func NewAutotesterController(
	logger *slog.Logger,
	config *config.Autotester,
	validationService service.Validator,
	generationService service.GeneratePrompt,
	localTestcaseStorageService service.TestcaseLocalStorageService,
	dockerService service.Docker,
	chatStorageService service.ChatStorageService,
	remoteTestcaseStorageService service.TestcaseStorageService,
	mediaStorageService service.MediaStorageService,
	chatManager service.ChatManager,
	groupManager service.GroupManager,
	tracer trace.Tracer,
	metricsService shared.MetricsService,
	authService service.Auth,
) (*AutotesterController, error) {
	if err := assert.NotNil(
		logger,
		config,
		validationService,
		generationService,
		localTestcaseStorageService,
		dockerService,
		remoteTestcaseStorageService,
		chatStorageService,
		mediaStorageService,
		chatManager,
		groupManager,
		tracer,
		metricsService,
		authService,
	); err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:                       logger,
		config:                       config,
		validationService:            validationService,
		generationService:            generationService,
		localTestcaseStorageService:  localTestcaseStorageService,
		dockerService:                dockerService,
		chatStorageService:           chatStorageService,
		remoteTestcaseStorageService: remoteTestcaseStorageService,
		mediaStorageService:          mediaStorageService,
		chatManager:                  chatManager,
		groupManager:                 groupManager,
		tracer:                       tracer,
		metricsService:               metricsService,
		authService:                  authService,
	}, nil
}
