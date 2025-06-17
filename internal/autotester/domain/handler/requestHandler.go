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
		return
	}

	// validation Method

	// Call Request to openAI
	// service needed
	// entity.Request(body.Message, body.ConversationID)

	// handling response openai
	service, err := service.NewService(logger, 5)
	if err != nil {
		return
	}
	// DTO to real entitys

	_, error := service.Request(ctx, entity.RequestForLlmDTO{UserPrompt: userRequest.Message.MessageBody, SystemPrompt: "Du bist ein hilfreicher Assistent", Model: "gpt-o4"})

	var resp entity.ResponseForUserDTO
	resp.ResponseText.Text = error.Error()
	// respons from LLM To frontend
	c.IndentedJSON(http.StatusOK, resp)
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
