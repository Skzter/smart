package main

import (
	"context"
	"log/slog"
	"os"

	config "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/configs"
	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}
	apiKey := config.GetOpenAIKey()

	repo := repository.NewOpenAi(apiKey)
	request := entity.Request{
		Model: "gpt-4o",
		Body: entity.RequestBody{
			UserPrompt: "Please generate test scenarios for my web application that validate core functionalities and edge cases.",
			SystemPrompt: "You are a testing assistant that specializes in creating comprehensive test cases for web applications." +
				"Focus on generating clear, specific, and actionable test scenarios that cover both happy paths and edge cases.",
		},
	}

	logger.Info("Sending request to OpenAI")
	resp, err := repo.CreateRequest(request, context.Background())

	if err != nil {
		logger.Error("Failed to create request", "error", err)
		return
	}
	logger.Info("Received response", "status", resp.Status, "id", resp.Id)
}
