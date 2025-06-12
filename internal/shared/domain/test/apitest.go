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
	loader := config.NewConfigLoader(logger)

	err := loader.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}
	apiKey := config.GetOpenAIKey()

	openAIService := service.NewService(
		repository.NewOpenAiRepository(logger, apiKey),
		"You are a helpful AI assistant. Please provide clear and concise responses.",
		"gpt-4.1",
		logger,
	)

	// Test conversation functionality
	logger.Info("Testing conversation")

	// First conversation message
	resp, err := openAIService.Request(service.NewRequest("Remember this fact: The speed of light is 299,792 kilometers per second."))
	if err != nil {
		logger.Error("Conversation first message failed", "error", err)
		return
	}
	logger.Info("First conversation response", "id", resp.Id, "text", resp.Text)
	println("First Response:", resp.Text)

	// Follow-up conversation message
	resp, err = openAIService.Request(service.NewRequestSession("Remember this fact: The speed of light is 299,792 kilometers per second.", resp.Id))
	if err != nil {
		logger.Error("Conversation follow-up failed", "error", err)
		return
	}
	logger.Info("Follow-up conversation response", "id", resp.Id, "text", resp.Text)
	println("Follow-up Response:", resp.Text)
}
