package repository

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

func ConnectAndRequestOpenAI(prompt, systemPrompt, model string) (*entity.Response, error) {
	logger := slog.New(slog.Default().Handler())
	timeout := 5 // seconds

	repo, err := repository.NewOpenAiRepository(logger, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI repository: %w", err)
	}

	req := entity.Request{
		Prompt:       prompt,
		SystemPrompt: systemPrompt,
		Model:        model,
	}

	ctx := context.Background()
	return repo.CreateRequest(ctx, req)
}
