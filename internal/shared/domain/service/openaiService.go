package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

type Service struct {
	api          repository.RepositoryInterface
	systemPrompt string
	model        string
	logger       *slog.Logger
}

type Conversation struct {
	lastID  string
	service *Service
}

func NewService(api repository.RepositoryInterface, systemPrompt string, model string, logger *slog.Logger) Service {
	return Service{api, systemPrompt, model, logger}
}

func (c *Service) RequestWithoutSession(prompt string) (string, error) {
	request := entity.NewRequest(c.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.systemPrompt})

	c.logger.Info("Sending simple request to OpenAI",
		"model", c.model,
		"prompt", prompt,
		"system_prompt", c.systemPrompt)

	resp, err := c.api.CreateRequest(context.Background(), request)
	if err != nil {
		return "", err
	}
	return resp.Text, err
}

func (c *Service) CreateConversation() Conversation {
	return Conversation{service: c}
}

func (c *Conversation) Request(prompt string) (string, error) {
	var resp entity.Response
	var err error
	if c.lastID == "" {
		request := entity.NewRequest(c.service.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.service.systemPrompt})
		c.service.logger.Info("Starting new conversation",
			"model", c.service.model,
			"prompt", prompt,
			"system_prompt", c.service.systemPrompt)
		resp, err = c.service.api.CreateRequest(context.Background(), request)
	} else {
		c.service.logger.Info("Continuing conversation",
			"model", c.service.model,
			"prompt", prompt,
			"system_prompt", c.service.systemPrompt,
			"previous_response_id", c.lastID)
		request := entity.NewRequestSession(c.service.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.service.systemPrompt}, c.lastID)

		resp, err = c.service.api.CreateRequest(context.Background(), request)
	}

	if err == nil {
		c.lastID = resp.Id
	}
	return resp.Text, err
}
