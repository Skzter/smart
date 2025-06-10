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

	service := service.NewService(
		repository.NewOpenAi(logger, apiKey),
		"You are a helpful AI assistant. Please provide clear and concise responses.",
		"gpt-4",
		logger,
	)

	// Test single request functionality
	logger.Info("Testing single request")
	resp, err := service.RequestWithoutSession("What is the capital of France?")
	if err != nil {
		logger.Error("Single request failed", "error", err)
		return
	}
	logger.Info("Single request response", "text", resp)
	println("Single Request Response:", resp)

	// Test conversation functionality
	logger.Info("Testing conversation")
	conv := service.CreateConversation()

	// First conversation message
	resp, err = conv.Request("Remember this fact: The speed of light is 299,792 kilometers per second.")
	if err != nil {
		logger.Error("Conversation first message failed", "error", err)
		return
	}
	logger.Info("First conversation response", "text", resp)
	println("First Response:", resp)

	// Follow-up conversation message
	resp, err = conv.Request("What fact did I just tell you about speed?")
	if err != nil {
		logger.Error("Conversation follow-up failed", "error", err)
		return
	}
	logger.Info("Follow-up conversation response", "text", resp)
	println("Follow-up Response:", resp)
}
