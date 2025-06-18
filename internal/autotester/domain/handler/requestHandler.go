package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	repoEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

type AutotesterController struct {
	logger  *slog.Logger
	service *service.OpenAIService
}

func NewAutotesterController(logger *slog.Logger) (a *AutotesterController, err error) {
	service, err := service.NewService(logger, 5)
	if err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:  logger,
		service: service,
	}, err
}

// handle Frontend Request JSON to String
func (a *AutotesterController) HandleChatRequest(c *gin.Context) {
	var userRequest entity.UserRequestDTO

	if err := c.BindJSON(&userRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("JSON binding failed", "error", err)
		return
	}

	// validation Method

	// Call Request to openAI
	// service needed
	// entity.Request(body.Message, body.ConversationID)
	resp, err := a.ServiceHandler(c, userRequest)
	if err != nil {
		a.logger.Error(err.Error())
		return
	}
	// respons from LLM To frontend
	c.IndentedJSON(http.StatusOK, resp)
}

func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequestDTO
	var resp entity.ResponseForUser
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.IndentedJSON(http.StatusOK, entity.ResponseForUser{LogStamp: resp.LogStamp, SessionId: resp.SessionId})
}

func (a *AutotesterController) ServiceHandler(c *gin.Context, userRequest entity.UserRequestDTO) (response *entity.ResponseForUserDTO, err error) {
	resp, err := a.service.Request(c, repoEntity.Request{
		Prompt:       userRequest.Message.MessageBody,
		SessionID:    userRequest.SessionId,
		SystemPrompt: "Du bist ein hilfreicher Assistent",
		Model:        "gpt-4.1-nano-2025-04-14"},
	)

	if err != nil {
		return nil, err
	}
	text := entity.MessageDTO{MessageBody: resp.Text, Actor: "system"}
	return &entity.ResponseForUserDTO{
		ResponseText: text,
		SessionIdDTO: entity.SessionIdDTO{Id: resp.SessionID},
		LogStampDTO:  entity.LogStampDTO{ActorId: ""},
	}, nil
}
