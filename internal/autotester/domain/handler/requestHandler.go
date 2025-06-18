package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// handle Frontend Request JSON to String
func HandleChatRequest(c *gin.Context, logger *slog.Logger, ctx context.Context) {
	var userRequest entity.UserRequestDTO

	if err := c.BindJSON(&userRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		logger.Error("JSON binding failed", "error", err)
		return
	}

	// validation Method

	// handling response openai
	service, err := service.NewService(logger, 5)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// DTO to real entitys

	resp, err := service.Request(ctx, entity.RequestForLlmDTO{
		UserPrompt:   userRequest.Message.MessageBody,
		SessionID:    userRequest.SessionId,
		SystemPrompt: "Du bist ein hilfreicher Assistent",
		Model:        "gpt-4.1-nano-2025-04-14"},
	)

	if err != nil {
		logger.Error(err.Error())
		return
	}
	text := entity.MessageDTO{MessageBody: resp.Text, Actor: "system"}
	response := entity.ResponseForUserDTO{
		ResponseText: text,
		SessionIdDTO: entity.SessionIdDTO{Id: resp.SessionID},
		LogStampDTO:  entity.LogStampDTO{ActorId: ""},
	}

	// respons from LLM To frontend
	c.IndentedJSON(http.StatusOK, response)
}

func HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequestDTO
	var resp entity.ResponseForUser
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.IndentedJSON(http.StatusOK, entity.ResponseForUser{LogStamp: resp.LogStamp, SessionId: resp.SessionId})
}
