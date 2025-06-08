package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

type Client struct {
	api          repository.RepositoryInterface
	systemPrompt string
	model        string
	logger       *slog.Logger
}

type Conversation struct {
	lastID string
	client *Client
}

func NewService(api repository.RepositoryInterface, systemPrompt string, model string, logger *slog.Logger) Client {
	return Client{api, systemPrompt, model, logger}
}

func (c *Client) RequestWithoutSession(prompt string) (entity.Response, error) {
	request := entity.NewRequest(c.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.systemPrompt})

	c.logger.Info("Sending simple request to OpenAI")
	return c.api.CreateRequest(request, context.Background(), c.logger)
}

func (c *Client) CreateConversation() Conversation {
	return Conversation{client: c}
}

func (c *Conversation) Request(prompt string) (entity.Response, error) {
	var resp entity.Response
	var err error
	if c.lastID == "" {
		c.client.logger.Info("No previous successful request, sending simple request")
		resp, err = c.client.RequestWithoutSession(prompt)
	} else {
		c.client.logger.Info("Sending conversation request to OpenAI")
		request := entity.NewRequestSession(c.client.model, entity.RequestBody{UserPrompt: prompt, SystemPrompt: c.client.systemPrompt}, c.lastID)

		resp, err = c.client.api.CreateRequest(request, context.Background(), c.client.logger)
	}

	if err == nil {
		c.lastID = resp.Id
	}
	return resp, err
}
