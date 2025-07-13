package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	repoEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterController is the controller for autotesting requests.
// It encapsulates logging and access to the OpenAI service.
type AutotesterController struct {
	config  *config.Config
	logger  *slog.Logger
	service *service.OpenAIService
}

// NewAutotesterController creates a new AutotesterController.
// Returns an initialized controller or an error.
func NewAutotesterController(logger *slog.Logger, config *config.Config, service *service.OpenAIService) (*AutotesterController, error) {
	if err := assert.NotNil(logger, config, service); err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:  logger,
		service: service,
		config:  config,
	}, nil
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

	resp, err := a.serviceHandler(c, userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "OpenAI service failed"})
		a.logger.Error(err.Error())
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
// Returns a ResponseForUserDTO or an error.

func (a *AutotesterController) serviceHandler(c *gin.Context, userRequest entity.UserRequest) (*entity.ResponseForUser, error) {
	resp, err := a.service.Request(c, repoEntity.Request{
		Prompt:       userRequest.Message.MessageBody,
		SessionID:    userRequest.SessionId,
		SystemPrompt: a.config.Prompts.ValidationPrompt,
		Model:        a.config.Model},
	)

	if err != nil {
		return nil, err
	}
	text := entity.Message{MessageBody: resp.Text, Actor: "system"}
	newLogStamp, err := entity.NewLogStamp(text.Actor)
	if err != nil {
		return nil, err
	}
	return &entity.ResponseForUser{
		Message:   text,
		UserId:    "",
		SessionId: resp.SessionID,
		LogStamp:  newLogStamp,
	}, nil
}
