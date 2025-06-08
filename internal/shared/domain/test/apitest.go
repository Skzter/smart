package main

import (
	"log/slog"
	"os"

	config "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/configs"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}
	apiKey := config.GetOpenAIKey()

	service := service.NewService(
		repository.NewOpenAi(apiKey),
		"This is a test, answer however you like to future Requests",
		"gpt-4o",
		logger,
	)

	// Simplte request test
	//
	// resp, err := service.RequestWithoutSession("Pleae tell me about diffren services Check24 offers")

	// Conversation test
	// logger.Info("Creating conversation")
	conv := service.CreateConversation()
	// logger.Info("Sending conversation request to OpenAI")
	// conv.Request("The secret number is 24. Remember this.")
	resp, err := conv.Request("what is the secret number?")

	/*
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
		resp, err := repo.CreateRequest(request, context.Background(), logger)
	*/

	if err != nil {
		return
	}
	logger.Info("Received response", "status", resp.Status, "id", resp.Id)
	println(resp.Output)
}
