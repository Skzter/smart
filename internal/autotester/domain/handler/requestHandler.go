package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

var (
	errOpenAIFailure = errors.New("failure in the request to open ai")
)

// AutotesterController is the controller for autotesting requests.
// It encapsulates logging and access to the OpenAI service.
type AutotesterController struct {
	config            *config.Config
	logger            *slog.Logger
	validationService service.ValidatePrompt
	generationService service.GeneratePrompt
}

// NewAutotesterController creates a new AutotesterController.
// Returns an initialized controller or an error.
func NewAutotesterController(
	logger *slog.Logger,
	config *config.Config,
	validationService service.ValidatePrompt,
	generationService service.GeneratePrompt,
) (*AutotesterController, error) {
	if err := assert.NotNil(logger, config, validationService, generationService); err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:            logger,
		config:            config,
		validationService: validationService,
		generationService: generationService,
	}, nil
}

// HandleGetTemplate processes a template request from the frontend.
func (a *AutotesterController) HandleGetTemplate(c *gin.Context) {
	if err := assert.StringNotEmpty(a.config.Template); err != nil {
		c.JSON(http.StatusTeapot, "")
		a.logger.Error(fmt.Sprintf("HANDLER: getTemplate() => %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, entity.Template{Template: a.config.Template})
}

// HandleChatRequest processes a chat request from the frontend.
// Expects a JSON with UserRequestDTO and returns a response from the LLM.
func (a *AutotesterController) HandleChatRequest(c *gin.Context) {
	var userRequest entity.UserRequest

	if err := c.BindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("JSON binding failed", "error", err)
		return
	}

	// returns handled errors which can be given to frontend
	resp, err := a.serviceHandler(c, userRequest)
	if err != nil {
		frontendError := handleError(err)
		unwrappedError := errors.Unwrap(frontendError)
		// checks if the error need the unwrapped
		// Unwrap() returns nil if nothing needs to be unwrapped
		if unwrappedError != nil {
			frontendError = unwrappedError
		}
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: frontendError.Error()})
		a.logger.Error(fmt.Sprintf("HANDLER serviceHandler: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleUserInfoRequest processes a request for user information.
// Expects a JSON with UserRequestDTO and returns a ResponseForUser.
func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequest
	var resp entity.ResponseForUser
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.JSON(http.StatusOK, entity.ResponseForUser{LogStamp: resp.LogStamp, SessionId: resp.SessionId})
}

// serviceHandler calls the OpenAI service and prepares the response for the frontend.
// It takes a gin.Context and UserRequest as input parameters.
// First validates the user prompt through validationService, then generates a response through generationService.
// Returns a ResponseForUser containing the generated text and user metadata, or an error if validation or generation fails.
func (a *AutotesterController) serviceHandler(c *gin.Context, userRequest entity.UserRequest) (*entity.ResponseForUser, error) {
	err := a.validationService.ValidatePrompt(c, userRequest.Message.MessageBody, userRequest.SessionId)
	if err != nil {
		return nil, err
	}
	genText, err := a.generationService.GeneratePrompt(c, userRequest.Message.MessageBody, userRequest.SessionId)
	if err != nil {
		return nil, err
	}
	return &entity.ResponseForUser{
		Message:   sharedEntity.Message{MessageBody: genText},
		UserId:    userRequest.UserId,
		SessionId: userRequest.SessionId,
		LogStamp:  userRequest.LogStamp,
	}, nil
}

// handleError handles the errors for the validate and generate functions
// takes the given error and checks if it is a validation or generation error and looks further if its from there or deeper in the system
// if source found decides if it can return this error or a generic error because the error message may expose sensitive data
func handleError(err error) error {
	if err == nil {
		return nil
	}
	// Generate() returns only the repository/service errors
	var (
		repoError         *repository.Error
		assertNotNilError *assert.NotNilError
		apiErr            *openai.APIError
		reqErr            *openai.RequestError
	)

	// maps the given error to the target error, when success you can operate on the custom error type
	switch {
	case errors.As(err, &repoError):
		switch repoError.Type {
		case repository.Private:
			return errOpenAIFailure
		case repository.Public:
			return repoError
		}
	case errors.As(err, &assertNotNilError),
		errors.As(err, &apiErr),
		errors.As(err, &reqErr):
		// catching errors from the assert and OpenAI requests
		return errOpenAIFailure
	case strings.Contains(err.Error(), "Post"):
		// error when no internet connection but bit unclean
		return errOpenAIFailure
	}

	// Validate() returns unique errors but they dont contain sensitive information
	return err
}
