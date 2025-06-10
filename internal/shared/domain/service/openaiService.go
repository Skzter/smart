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
	lastID string
	client *Service
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
	return resp.Output, err
}

func (c *Service) CreateConversation() Conversation {
	return Conversation{client: c}
}

func (c *Conversation) Request(prompt string) (string, error) {
	var resp entity.Response
	var err error
	if c.lastID == "" {
		request := entity.NewRequest(c.client.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.client.systemPrompt})
		c.client.logger.Info("Starting new conversation",
			"model", c.client.model,
			"prompt", prompt,
			"system_prompt", c.client.systemPrompt)
		resp, err = c.client.api.CreateRequest(context.Background(), request)
	} else {
		c.client.logger.Info("Continuing conversation",
			"model", c.client.model,
			"prompt", prompt,
			"system_prompt", c.client.systemPrompt,
			"previous_response_id", c.lastID)
		request := entity.NewRequestSession(c.client.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.client.systemPrompt}, c.lastID)

		resp, err = c.client.api.CreateRequest(context.Background(), request)
	}

	if err == nil {
		c.lastID = resp.Id
	}
	return resp.Output, err
}
